import pandas as pd
import hashlib
import time
from typing import List, Tuple, Set
import random


class MerkleTree:
    def __init__(self, chunk_size: int):
        self.chunk_size = chunk_size
        self.leaves = []  # 存储叶子节点哈希
        self.tree = []  # 存储整棵树的所有层
        self.proofs = {}  # 存储每个叶子节点的预生成证明，键为块索引

    def hash_data(self, data: str) -> str:
        return hashlib.sha256(str(data).encode()).hexdigest()

    def create_leaf_nodes(self, df: pd.DataFrame) -> List[str]:
        leaves = []
        for i in range(0, len(df), self.chunk_size):
            chunk = df.iloc[i:i + self.chunk_size]
            chunk_str = ''.join(chunk.astype(str).values.flatten())
            leaves.append(self.hash_data(chunk_str))
        self.leaves = leaves
        return leaves

    def build_tree(self, df: pd.DataFrame) -> str:
        self.create_leaf_nodes(df)

        if not self.leaves:
            raise ValueError("输入数据为空，无法构建 Merkle 树")

        current_level = self.leaves
        self.tree = [current_level]

        while len(current_level) > 1:
            next_level = []
            for i in range(0, len(current_level), 2):
                left = current_level[i]
                right = current_level[i + 1] if i + 1 < len(current_level) else left
                combined_hash = self.hash_data(left + right)
                next_level.append(combined_hash)
            current_level = next_level
            self.tree.append(current_level)

        # 构建并存储每个叶子节点的证明
        root_hash = current_level[0]
        for i in range(len(self.leaves)):
            self.proofs[i] = self.get_proof(i)  # 预生成每个块的证明

        return root_hash

    def get_proof(self, chunk_index: int) -> List[Tuple[str, bool]]:
        proof = []
        current_index = chunk_index
        for level in self.tree[:-1]:
            pair_index = current_index ^ 1
            if pair_index >= len(level):
                pair_index = current_index
            is_left = (current_index % 2 == 0)
            proof.append((level[pair_index], is_left))
            current_index = current_index // 2
        return proof

    def verify_proof(self, leaf_hash: str, proof: List[Tuple[str, bool]], root_hash: str) -> bool:
        current_hash = leaf_hash
        print("当前叶子节点哈希值：", leaf_hash)
        print("当前Merkle证明:", proof)
        print("当前根hash", root_hash)
        for node_hash, is_left in proof:
            if is_left:
                current_hash = self.hash_data(current_hash + node_hash)
            else:
                current_hash = self.hash_data(node_hash + current_hash)
        print("根据Merkle证明计算当前hash",current_hash)
        flag = (current_hash == root_hash)
        print("验证计算hash和根hash是否一致：", flag)
        print()
        return flag


# 其他函数保持不变（generate_simulated_data, analyze_results 等）

def generate_simulated_data(data_size: int) -> pd.DataFrame:
    data = {
        'AGE': [f"AGE_{random.randint(20, 60)}" for _ in range(data_size)],
        'BILLAMT1': [round(random.uniform(0, 100000), 6) for _ in range(data_size)],
        'BILLAMT2': [round(random.uniform(0, 100000), 6) for _ in range(data_size)],
        'BILLAMT3': [round(random.uniform(0, 100000), 6) for _ in range(data_size)],
        'BILLAMT4': [round(random.uniform(0, 100000), 6) for _ in range(data_size)],
        'BILLAMT5': [round(random.uniform(0, 100000), 6) for _ in range(data_size)],
        'BILLAMT6': [round(random.uniform(0, 100000), 6) for _ in range(data_size)],
        'EDUCATION': [f"EDUCATION_{random.randint(1, 6)}" for _ in range(data_size)],
        'LIMITBAL': [random.randint(10000, 100000) for _ in range(data_size)],
        'MARRIAGE': [f"MARRIAGE_{random.randint(1, 3)}" for _ in range(data_size)],
        'PAY0': [f"PAY0_{random.randint(-1, 8)}" for _ in range(data_size)],
        'PAY2': [f"PAY2_{random.randint(-1, 8)}" for _ in range(data_size)],
        'PAY3': [f"PAY3_{random.randint(-1, 8)}" for _ in range(data_size)],
        'PAY4': [f"PAY4_{random.randint(-1, 8)}" for _ in range(data_size)],
        'PAY5': [f"PAY5_{random.randint(-1, 8)}" for _ in range(data_size)],
        'PAY6': [f"PAY6_{random.randint(-1, 8)}" for _ in range(data_size)],
        'PAYAMT1': [round(random.uniform(0, 10000), 2) for _ in range(data_size)],
        'PAYAMT2': [round(random.uniform(0, 10000), 2) for _ in range(data_size)],
        'PAYAMT3': [round(random.uniform(0, 10000), 2) for _ in range(data_size)],
        'PAYAMT4': [round(random.uniform(0, 10000), 2) for _ in range(data_size)],
        'PAYAMT5': [round(random.uniform(0, 10000), 2) for _ in range(data_size)],
        'PAYAMT6': [round(random.uniform(0, 10000), 2) for _ in range(data_size)],
        'SEX': [f"SEX_{random.randint(1, 2)}" for _ in range(data_size)]
    }
    return pd.DataFrame(data)


def run_experiment(data_size: int, chunk_sizes: List[int], num_modifications: int = 1) -> pd.DataFrame:
    df = generate_simulated_data(data_size)
    results = []

    max_metrics = {'build_time': 0, 'validation_time': 0, 'discarded_ratio': 0}

    for chunk_size in chunk_sizes:
        tree = MerkleTree(chunk_size)

        start = time.perf_counter()
        root = tree.build_tree(df)
        build_time = time.perf_counter() - start

        start = time.perf_counter()
        validation_success = True
        num_validations = 100
        num_chunks_to_validate = min(10, len(tree.leaves))

        for _ in range(num_validations):
            chunk_indices = random.sample(range(len(tree.leaves)), num_chunks_to_validate)
            for chunk_index in chunk_indices:
                chunk_hash = tree.leaves[chunk_index]
                # 使用预生成的证明
                proof = tree.proofs[chunk_index]  # 直接从预生成证明中获取
                if not tree.verify_proof(chunk_hash, proof, root):
                    validation_success = False
                    break
            if not validation_success:
                break

        validation_time = (time.perf_counter() - start) / num_validations

        modified_indices = random.sample(range(len(df)), num_modifications)
        affected_chunks = set()
        for idx in modified_indices:
            chunk_idx = idx // chunk_size
            affected_chunks.add(chunk_idx)
            if idx % chunk_size == 0 and chunk_idx > 0:
                affected_chunks.add(chunk_idx - 1)
            if idx % chunk_size == chunk_size - 1 and chunk_idx < (len(df) + chunk_size - 1) // chunk_size - 1:
                affected_chunks.add(chunk_idx + 1)

        total_chunks = (len(df) + chunk_size - 1) // chunk_size
        discarded_ratio = len(affected_chunks) / total_chunks

        max_metrics['build_time'] = max(max_metrics['build_time'], build_time)
        max_metrics['validation_time'] = max(max_metrics['validation_time'], validation_time)
        max_metrics['discarded_ratio'] = max(max_metrics['discarded_ratio'], discarded_ratio)

    for chunk_size in chunk_sizes:
        tree = MerkleTree(chunk_size)

        start = time.perf_counter()
        root = tree.build_tree(df)
        build_time = time.perf_counter() - start

        start = time.perf_counter()
        validation_success = True
        num_validations = 100
        num_chunks_to_validate = min(10, len(tree.leaves))

        for _ in range(num_validations):
            chunk_indices = random.sample(range(len(tree.leaves)), num_chunks_to_validate)
            for chunk_index in chunk_indices:
                chunk_hash = tree.leaves[chunk_index]
                proof = tree.proofs[chunk_index]  # 直接使用预生成的证明
                if not tree.verify_proof(chunk_hash, proof, root):
                    validation_success = False
                    break
            if not validation_success:
                break

        validation_time = (time.perf_counter() - start) / num_validations

        modified_indices = random.sample(range(len(df)), num_modifications)
        affected_chunks = set()
        for idx in modified_indices:
            chunk_idx = idx // chunk_size
            affected_chunks.add(chunk_idx)
            if idx % chunk_size == 0 and chunk_idx > 0:
                affected_chunks.add(chunk_idx - 1)
            if idx % chunk_size == chunk_size - 1 and chunk_idx < (len(df) + chunk_size - 1) // chunk_size - 1:
                affected_chunks.add(chunk_idx + 1)

        total_chunks = (len(df) + chunk_size - 1) // chunk_size
        discarded_ratio = len(affected_chunks) / total_chunks

        norm_build = build_time / max_metrics['build_time'] if max_metrics['build_time'] > 0 else 0
        norm_valid = validation_time / max_metrics['validation_time'] if max_metrics['validation_time'] > 0 else 0
        norm_discard = discarded_ratio / max_metrics['discarded_ratio'] if max_metrics['discarded_ratio'] > 0 else 0
        score = norm_build + norm_valid + norm_discard

        results.append({
            'chunk_size': chunk_size,
            'num_leaves': len(tree.leaves),
            'build_time': round(build_time, 6),
            'validation_time': round(validation_time, 6),
            'discarded_ratio': round(discarded_ratio, 6),
            'score': round(score, 6)
        })

    return pd.DataFrame(results)


def analyze_results(results: pd.DataFrame) -> None:
    pass
    # print("\n实验结果分析:")
    # print("-" * 50)
    # print("\n完整结果表格:")
    # print(results.to_string(float_format=lambda x: '{:.6f}'.format(x) if isinstance(x, float) else str(x)))

    # print("\n关键指标:")
    # print(f"构建时间范围: {results['build_time'].min():.6f}s 到 {results['build_time'].max():.6f}s")
    # print(f"验证时间范围: {results['validation_time'].min():.6f}s 到 {results['validation_time'].max():.6f}s")
    # print(f"抛弃比例范围: {results['discarded_ratio'].min():.6f} 到 {results['discarded_ratio'].max():.6f}")

    # best_chunk = results.loc[results['score'].idxmin()]

    # print(f"\n推荐的chunk大小: {int(best_chunk['chunk_size'])}")
    # print(f"- 构建时间: {best_chunk['build_time']:.6f}s")
    # print(f"- 验证时间: {best_chunk['validation_time']:.6f}s")
    # print(f"- 抛弃比例: {best_chunk['discarded_ratio']:.6f}")
    # print(f"- 叶子节点数量: {int(best_chunk['num_leaves'])}")
    # print(f"- 综合分数: {best_chunk['score']:.6f}")


if __name__ == "__main__":
    data_size = 100000
    # chunk_sizes = [1, 5, 10, 25, 50, 75, 100, 150, 200, 300, 400, 500, 750, 1000]
    chunk_sizes = [400, 500, 750, 1000]
    results = run_experiment(data_size, chunk_sizes)
    analyze_results(results)
