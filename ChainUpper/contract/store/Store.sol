// SPDX-License-Identifier: MIT
pragma solidity ^0.8.11;

contract Store {
    event ItemSet(bytes32 key, bytes32 value);

    mapping (bytes32 => bytes32) public items;

    function setItem(bytes32 key, bytes32 value) external {
        items[key] = value;
        emit ItemSet(key, value);
    }

    function setItems(bytes32[] calldata keys, bytes32[] calldata values) external {
        require(keys.length == values.length, "Keys and values array must be of same length");
        
        for (uint256 i = 0; i < keys.length; i++) {
            items[keys[i]] = values[i];
            emit ItemSet(keys[i], values[i]);
        }
    }

    function getItem(bytes32 key) external view returns (bytes32) {
        return items[key];
    }
}
