package verkle

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"
)

// RunVerkleExperiment 输出详细的构建和验证过程，展示证明的所有细节
func RunVerkleExperiment(dataSize int) {
	// 1. 生成模拟数据
	data := make([][]byte, dataSize)
	for i := 0; i < dataSize; i++ {
		key := make([]byte, 32)
		rand.Read(key)
		data[i] = key
	}

	vt := &VerkleTree{}

	fmt.Println("==================================================")
	fmt.Printf("Verkle Tree 全量验证实验开始\n")
	fmt.Printf("展示所有数据、证明细节和承诺\n")
	fmt.Printf("数据规模: %d\n", dataSize)
	fmt.Println("==================================================")

	// 2. 构建树
	fmt.Println("[1/3] 开始构建树...")
	start := time.Now()
	err := vt.Build(data)
	if err != nil {
		fmt.Printf("构建失败: %v\n", err)
		return
	}
	buildTime := time.Since(start)
	fmt.Printf("构建完成！耗时: %v\n", buildTime)

	// 3. 全量验证输出
	fmt.Printf("\n[2/3] 开始遍历验证所有节点 (%d 个):\n", dataSize)
	for idx := 0; idx < dataSize; idx++ {
		targetHash := data[idx]

		fmt.Printf("\n##################################################\n")
		fmt.Printf("## 节点索引: %d\n", idx)
		fmt.Printf("## 原始数据哈希 (Hash): %x\n", targetHash)

		// 获取证明
		proofData, ok := vt.GetProof(idx)
		if !ok {
			fmt.Printf("## 错误: 无法为索引 %d 生成证明\n", idx)
			continue
		}

		serialized := proofData.(VerkleSerializedProof)
		vp := serialized.Vp

		// 展示证明的所有详细内容
		fmt.Printf("## --- Verkle 证明 (Proof) 详细信息 ---\n")
		fmt.Printf("## 域承诺 (D - Domain Commitment): %x\n", vp.D)

		fmt.Printf("## 其它 Stems (存在/缺失证明): %d 个\n", len(vp.OtherStems))
		for j, stem := range vp.OtherStems {
			fmt.Printf("##   位置 [%d]: %x\n", j, stem)
		}

		fmt.Printf("## 路径承诺 (Commitments By Path): %d 个\n", len(vp.CommitmentsByPath))
		for j, comm := range vp.CommitmentsByPath {
			fmt.Printf("##   深度 [%d]: %x\n", j, comm)
		}

		fmt.Printf("## 深度扩展标志 (Depth Extension Present): %x\n", vp.DepthExtensionPresent)

		if vp.IPAProof != nil {
			fmt.Printf("## IPA 证明细节:\n")
			fmt.Printf("##   最终评估值 (Final Evaluation): %x\n", vp.IPAProof.FinalEvaluation)
			fmt.Printf("##   CL 向量 (左侧向量，深度 %d):\n", len(vp.IPAProof.CL))
			for j, cl := range vp.IPAProof.CL {
				fmt.Printf("##     [%d]: %x\n", j, cl)
			}
			fmt.Printf("##   CR 向量 (右侧向量，深度 %d):\n", len(vp.IPAProof.CR))
			for j, cr := range vp.IPAProof.CR {
				fmt.Printf("##     [%d]: %x\n", j, cr)
			}
		}

		// 根承诺 (Root Commitment)
		rootComm := vt.root.Commit()
		fmt.Printf("## 根承诺 (Root Commitment): %x\n", rootComm.Bytes())

		// 执行验证
		startVerify := time.Now()
		isValid := vt.Verify(targetHash, proofData)
		verifyTime := time.Since(startVerify)

		fmt.Printf("## 最终验证结果: %v (验证耗时: %v)\n", isValid, verifyTime)
		fmt.Printf("##################################################\n")
	}

	// 4. 统计总结
	fmt.Printf("\n[3/3] 实验总结:\n")
	fmt.Printf("- 总节点数: %d\n", dataSize)
	fmt.Printf("- 构建总时间: %v\n", buildTime)
	fmt.Println("==================================================")
}

// TestVerkleExperiment 一个函数展示全部逻辑
func TestVerkleExperiment(t *testing.T) {
	RunVerkleExperiment(100)
}
