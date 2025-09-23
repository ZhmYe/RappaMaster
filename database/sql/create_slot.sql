INSERT INTO slot (slotHash, taskID, nodeID, expected, scheduleEpoch)
VALUES (?,?,?,?,
    (SELECT IFNULL(MAX(id), -1) as current_epoch FROM epoch)
);