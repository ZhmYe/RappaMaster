/* 将当前所有的已经commit但是还没有justified的slot取出，用于初始化epochTree */
SELECT task.sign as task_sign, slot.nodeID as node_id, slot.commitment, slot.slotHash
FROM slot, task
WHERE slot.status = 'committed' and slot.taskID = task.id;