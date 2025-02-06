// SPDX-License-Identifier: MIT
pragma solidity ^0.8.11;

contract Store {
    event ItemSet(bytes32 indexed key, bytes value);

    // 使用 mapping 存储上链数据，key 为固定 32 字节，value 为动态字节数组
    mapping (bytes32 => bytes) public items;

    function setItem(bytes32 key, bytes calldata value) external {
        items[key] = value;
        emit ItemSet(key, value);
    }

    // 批量设置函数：接收 keys 数组和 values 数组，要求长度相同，
    function setItems(bytes32[] calldata keys, bytes[] calldata values) external {
        require(keys.length == values.length, "Keys and values array must be of same length");
        for (uint256 i = 0; i < keys.length; i++) {
            items[keys[i]] = values[i];
            emit ItemSet(keys[i], values[i]);
        }
    }

    // 读取函数：根据 key 查询对应的上链数据
    function getItem(bytes32 key) external view returns (bytes memory) {
        return items[key];
    }
}
