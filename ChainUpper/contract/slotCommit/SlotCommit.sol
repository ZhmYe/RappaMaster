// SPDX-License-Identifier: MIT
pragma solidity ^0.8.11;

contract SlotCommit {
    struct CommitSlotItem {
        bytes32 sign;    // 任务标识
        uint256 slot;    // 插槽信息
        uint256 process; // 节点处理进度
        uint256 nid;     // 节点 ID
        uint256 epoch;   // 当前 Epoch
    }

    // 将任务标识映射到多个 CommitSlotItem
    mapping(bytes32 => CommitSlotItem[]) public slotCommits;

    // 用于快速检测是否存在完全相同的 CommitSlotItem
    mapping(bytes32 => mapping(bytes32 => bool)) private itemExists;

    // 事件，用于链上通知
    event SlotCommitted(bytes32 indexed sign, CommitSlotItem item);
    event BatchSlotsCommitted(uint256 count);

    /// @notice 提交插槽信息
    /// @param sign 唯一任务标识
    /// @param slot 插槽信息
    /// @param process 节点处理进度
    /// @param nid 节点 ID
    /// @param epoch 当前 Epoch
    function commitSlot(
        bytes32 sign,
        uint256 slot,
        uint256 process,
        uint256 nid,
        uint256 epoch
    ) public {
        require(sign != bytes32(0), "Sign cannot be empty");

        // 计算 CommitSlotItem 的哈希值
        bytes32 itemHash = keccak256(abi.encodePacked(slot, process, nid, epoch));
        require(!itemExists[sign][itemHash], "Duplicate data: identical CommitSlotItem already exists");

        // 插入新的 CommitSlotItem
        CommitSlotItem memory item = CommitSlotItem({
            sign: sign,
            slot: slot,
            process: process,
            nid: nid,
            epoch: epoch
        });

        slotCommits[sign].push(item);
        itemExists[sign][itemHash] = true; // 标记为已存在

        emit SlotCommitted(sign, item);
    }

    /// @notice 批量提交插槽信息
    /// @param signs 任务标识数组
    /// @param slots 插槽信息数组
    /// @param processes 节点处理进度数组
    /// @param nids 节点 ID 数组
    /// @param epochs 当前 Epoch 数组
    function commitSlotsBatch(
        bytes32[] memory signs,
        uint256[] memory slots,
        uint256[] memory processes,
        uint256[] memory nids,
        uint256[] memory epochs
    ) public {
        require(
            signs.length == slots.length &&
            slots.length == processes.length &&
            processes.length == nids.length &&
            nids.length == epochs.length,
            "Input array lengths must match"
        );

        uint256 count = signs.length;

        for (uint256 i = 0; i < count; i++) {
            bytes32 sign = signs[i];
            uint256 slot = slots[i];
            uint256 process = processes[i];
            uint256 nid = nids[i];
            uint256 epoch = epochs[i];

            require(sign != bytes32(0), "Sign cannot be empty");

            // 计算 CommitSlotItem 的哈希值
            bytes32 itemHash = keccak256(abi.encodePacked(slot, process, nid, epoch));
            require(!itemExists[sign][itemHash], "Duplicate data in batch: identical CommitSlotItem already exists");

            // 插入新的 CommitSlotItem
            CommitSlotItem memory item = CommitSlotItem({
                sign: sign,
                slot: slot,
                process: process,
                nid: nid,
                epoch: epoch
            });

            slotCommits[sign].push(item);
            itemExists[sign][itemHash] = true; // 标记为已存在

            emit SlotCommitted(sign, item);
        }

        emit BatchSlotsCommitted(count);
    }

    /// @notice 查询特定任务标识的提交信息
    /// @param sign 任务标识
    /// @return CommitSlotItem[] 返回该任务标识的所有提交信息
    function getSlotCommits(bytes32 sign)
        public
        view
        returns (CommitSlotItem[] memory)
    {
        return slotCommits[sign];
    }
}
