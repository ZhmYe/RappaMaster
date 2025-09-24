package Grpc

import (
	"RappaMaster/crypto"
	"RappaMaster/epoch"
	"RappaMaster/helper"
	"RappaMaster/paradigm"
	pb "RappaMaster/pb/service"
	"context"
	"fmt"
)

// EpochValidationRequest represents epoch validation request from nodes
type EpochValidationRequest struct {
	EpochID      int32             `json:"epochID"`
	EpochRoot    string            `json:"epochRoot"`
	NodeSignatures map[int32]string `json:"nodeSignatures"` // nodeID -> BLS signature
}

// EpochValidationResponse represents response to epoch validation
type EpochValidationResponse struct {
	Success       bool     `json:"success"`
	ValidatedNodes []int32  `json:"validatedNodes"`
	JustifiedSlots []string `json:"justifiedSlots"`
}

// ValidateEpoch handles epoch validation with BLS signatures
func (ge *GrpcEngine) ValidateEpoch(ctx context.Context, req *EpochValidationRequest) (*EpochValidationResponse, error) {
	epochProcessor := epoch.NewEpochProcessor(helper.GlobalServiceHelper.DB)
	
	// Convert map[int32]string to map[int]string
	nodeSignatures := make(map[int]string)
	for nodeID, sig := range req.NodeSignatures {
		nodeSignatures[int(nodeID)] = sig
	}
	
	// Justify slots based on valid BLS signatures
	justifiedSlots, err := epochProcessor.JustifySlots(int(req.EpochID), nodeSignatures)
	if err != nil {
		paradigm.Log("ERROR", fmt.Sprintf("Failed to justify slots for epoch %d: %v", req.EpochID, err))
		return &EpochValidationResponse{
			Success: false,
		}, err
	}
	
	// Get list of validated nodes
	var validatedNodes []int32
	for nodeID := range req.NodeSignatures {
		validatedNodes = append(validatedNodes, nodeID)
	}
	
	paradigm.Log("INFO", fmt.Sprintf("Epoch %d validated: %d nodes, %d justified slots", 
		req.EpochID, len(validatedNodes), len(justifiedSlots)))
	
	return &EpochValidationResponse{
		Success:        true,
		ValidatedNodes: validatedNodes,
		JustifiedSlots: justifiedSlots,
	}, nil
}

// GetEpochMerkleProof returns merkle proof for a specific slot or task
func (ge *GrpcEngine) GetEpochMerkleProof(ctx context.Context, req *pb.MerkleProofRequest) (*pb.MerkleProofResponse, error) {
	// Query database for merkle proof
	var proof string
	var err error
	
	if req.SlotHash != "" {
		// Get slot merkle proof
		proof, err = ge.getSlotMerkleProof(req.SlotHash)
	} else if req.TaskSign != "" {
		// Get task merkle proof
		proof, err = ge.getTaskMerkleProof(req.TaskSign)
	} else {
		return &pb.MerkleProofResponse{
			Success: false,
			Error:   "Either slotHash or taskSign must be provided",
		}, fmt.Errorf("invalid request parameters")
	}
	
	if err != nil {
		return &pb.MerkleProofResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}
	
	return &pb.MerkleProofResponse{
		Success: true,
		Proof:   proof,
	}, nil
}

// getSlotMerkleProof retrieves merkle proof for a slot
func (ge *GrpcEngine) getSlotMerkleProof(slotHash string) (string, error) {
	query := "SELECT merkleProof FROM slot WHERE slotHash = ?"
	var proof string
	err := helper.GlobalServiceHelper.DB.QueryRow(query, slotHash).Scan(&proof)
	return proof, err
}

// getTaskMerkleProof retrieves task merkle proof for a slot
func (ge *GrpcEngine) getTaskMerkleProof(taskSign string) (string, error) {
	query := `
		SELECT s.taskMerkleProof 
		FROM slot s 
		JOIN task t ON s.taskID = t.id 
		WHERE t.sign = ? AND s.taskMerkleProof IS NOT NULL 
		LIMIT 1
	`
	var proof string
	err := helper.GlobalServiceHelper.DB.QueryRow(query, taskSign).Scan(&proof)
	return proof, err
}

// RegisterNodePublicKey registers a node's BLS public key
func (ge *GrpcEngine) RegisterNodePublicKey(ctx context.Context, req *pb.NodeKeyRegistrationRequest) (*pb.NodeKeyRegistrationResponse, error) {
	// Parse BLS public key
	publicKey, err := crypto.ParseBLSPublicKey(req.PublicKey)
	if err != nil {
		return &pb.NodeKeyRegistrationResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid public key: %v", err),
		}, err
	}
	
	// Store public key (in a real implementation, this would be stored in database)
	// For now, we'll use a global signature manager
	sigManager := crypto.NewSignatureManager()
	sigManager.RegisterNodePublicKey(int(req.NodeId), publicKey)
	
	paradigm.Log("INFO", fmt.Sprintf("Registered BLS public key for node %d", req.NodeId))
	
	return &pb.NodeKeyRegistrationResponse{
		Success: true,
	}, nil
}

// GetEpochInfo returns information about a specific epoch
func (ge *GrpcEngine) GetEpochInfo(ctx context.Context, req *pb.EpochInfoRequest) (*pb.EpochInfoResponse, error) {
	epochID := int(req.EpochId)
	
	// Get epoch root
	epochRoot, err := helper.GlobalServiceHelper.DB.GetEpochRoot(epochID)
	if err != nil {
		return &pb.EpochInfoResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get epoch root: %v", err),
		}, err
	}
	
	// Get committed slots count
	committedSlots, err := helper.GlobalServiceHelper.DB.GetCommittedSlotsInEpoch(epochID)
	if err != nil {
		return &pb.EpochInfoResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get committed slots: %v", err),
		}, err
	}
	
	return &pb.EpochInfoResponse{
		Success:        true,
		EpochId:        req.EpochId,
		EpochRoot:      epochRoot,
		CommittedSlots: int32(len(committedSlots)),
	}, nil
}