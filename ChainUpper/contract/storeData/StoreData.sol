// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

contract StoreData {
    // 1. InitTask
    struct InitTask {
        string    name;
        uint32    size;
        bytes32   model;
        bool      isReliable;
        bytes     params;
    }
    mapping(bytes32 => InitTask) private _initTasks;
    event InitTaskStored(
        bytes32 indexed sign,
        string    name,
        uint32    size,
        bytes32   model,
        bool      isReliable,
        bytes     params
    );

    function storeInitTask(
        bytes32 sign,
        string calldata name,
        uint32  size,
        bytes32 model,
        bool    isReliable,
        bytes   calldata params
    ) external {
        _initTasks[sign] = InitTask({
            name:       name,
            size:       size,
            model:      model,
            isReliable: isReliable,
            params:     params
        });
        emit InitTaskStored(sign, name, size, model, isReliable, params);
    }

    function getInitTask(bytes32 sign)
        external
        view
        returns (
            string memory name,
            uint32    size,
            bytes32   model,
            bool      isReliable,
            bytes memory params
        )
    {
        InitTask storage t = _initTasks[sign];
        return (t.name, t.size, t.model, t.isReliable, t.params);
    }

    // 2. TaskProcess
    struct TaskProcess {
        bytes32 sign;
        uint32 slot;
        uint32 process;
        uint32 id;
        uint32 epoch;
        bytes32 hash;
        bytes32 commitment;
        bytes     proof;
        bytes[]   signatures;
    }
    mapping(bytes32 => TaskProcess) private _taskProcesses;
    event TaskProcessStored(
        bytes32 indexed sign,
        bytes32 indexed hash,
        uint32        slot,
        uint32        process,
        uint32        id,
        uint32        epoch,
        bytes32       commitment,
        bytes         proof,
        bytes[]       signatures
    );

    function storeTaskProcess(
        bytes32 sign,
        bytes32 hash,
        uint32  slot,
        uint32  process,
        uint32  id,
        uint32  epoch,
        bytes32 commitment,
        bytes   calldata proof,
        bytes[] calldata signatures
    ) external {
        _taskProcesses[hash] = TaskProcess({
            sign:       sign,
            slot:       slot,
            process:    process,
            id:         id,
            epoch:      epoch,
            hash:       hash,
            commitment: commitment,
            proof:      proof,
            signatures: signatures
        });
        emit TaskProcessStored(sign, hash, slot, process, id, epoch, commitment, proof, signatures);
    }

    function getTaskProcess(bytes32 hash)
        external
        view
        returns (
            bytes32 sign,
            uint32  slot,
            uint32  process,
            uint32  id,
            uint32  epoch,
            bytes32 commitment,
            bytes   memory proof,
            bytes[] memory signatures
        )
    {
        TaskProcess storage t = _taskProcesses[hash];
        return (t.sign, t.slot, t.process, t.id, t.epoch, t.commitment, t.proof, t.signatures);
    }

    // 3. EpochRecord
    struct InvalidEntry {
        bytes32 hash;
        uint8   reason;
    }
    struct EpochRecord {
        bytes32[]      justified;
        bytes32[]      commits;
        InvalidEntry[] invalids;
    }
    mapping(uint256 => EpochRecord) private _epochRecords;
    event EpochRecordStored(
        uint256 indexed id,
        bytes32[] justified,
        bytes32[] commits,
        bytes32[] invalidHashes,
        uint8[]   invalidReasons
    );

    function storeEpochRecord(
        uint256       id,
        bytes32[] calldata justified,
        bytes32[] calldata commits,
        bytes32[] calldata invalidHashes,
        uint8[]   calldata invalidReasons
    ) external {
        require(invalidHashes.length == invalidReasons.length, "Mismatched invalid entries");
        EpochRecord storage r = _epochRecords[id];
        r.justified = justified;
        r.commits = commits;
        delete r.invalids;
        for (uint i = 0; i < invalidHashes.length; i++) {
            r.invalids.push(InvalidEntry({hash: invalidHashes[i], reason: invalidReasons[i]}));
        }
        emit EpochRecordStored(id, justified, commits, invalidHashes, invalidReasons);
    }

    /// @notice 按 id 查询 EpochRecord 所有信息
    function getEpochRecordFull(uint256 id)
        external
        view
        returns (
            bytes32[] memory justified,
            bytes32[] memory commits,
            InvalidEntry[] memory invalids
        )
    {
        EpochRecord storage r = _epochRecords[id];
        return (r.justified, r.commits, r.invalids);
    }
}
