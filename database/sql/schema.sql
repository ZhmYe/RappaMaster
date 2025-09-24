CREATE TABLE IF NOT EXISTS epoch (
    id INT PRIMARY KEY AUTO_INCREMENT,
    startDate Date NOT NULL,
    finishDate DATE,
    epochRoot VARCHAR(64)
) ENGINE=InnoDB DEFAULT CHARSET =utf8mb4;

-- create task table to record the tasks submitted by users in frontend
CREATE TABLE IF NOT EXISTS task (
    id INT PRIMARY KEY AUTO_INCREMENT, /* so that we don't need to write an id-allocator */
    sign VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL, /*  so we should limit the length of task */
    epochID INT NOT NULL, /* task created in which epoch(after tx is on chain)*/
    expected BIGINT NOT NULL,
    finished BIGINT NOT NULL DEFAULT 0,
    model VARCHAR(50) NOT NULL, /* we not use ENUM or SET since we want to have a dynamic set */
    txHash VARCHAR(64) NOT NULL,  /* transaction hash in blockchain, we must commit the transaction first */
    startDate DATE NOT NULL,
    finishDate DATE,
    done BOOL DEFAULT FALSE, /* has been finished */
    taskRoot VARCHAR(64), /* merkle root of all justified slots for this task */
    finishEpoch INT, /* epoch when task is completed */
    CONSTRAINT TaskCreateEpoch FOREIGN KEY (epochID) REFERENCES epoch(id),
    CONSTRAINT TaskFinishEpoch FOREIGN KEY (finishEpoch) REFERENCES epoch(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS slot (
    id INT PRIMARY KEY AUTO_INCREMENT,
    slotHash VARCHAR(64) NOT NULL,
    taskID INT NOT NULL,
    nodeID INT NOT NULL, /* commit by which node */
    expected BIGINT NOT NULL,
    finished BIGINT NOT NULL DEFAULT 0,
    scheduleEpoch INT NOT NULL,
    commitEpoch INT, /* commit but not finalized */
    finalizeEpoch INT,
    commitment VARCHAR(64), /* we not constrain not null here, we check in code */
    signature VARCHAR(64), 
    padding INT NOT NULL DEFAULT 0,
    store ENUM('local', 'ec') DEFAULT 'local',
    status ENUM('committed', 'justified', 'finalized') DEFAULT 'committed',
    merkleProof JSON, /* merkle proof for this slot within task */
    taskMerkleProof JSON, /* merkle proof for task within epoch */
    blsSignature VARCHAR(256), /* BLS signature from node */
    CONSTRAINT TaskCommitSlot FOREIGN KEY (taskID) REFERENCES task(id),
    CONSTRAINT SlotCommitEpoch FOREIGN KEY (commitEpoch) REFERENCES epoch(id),
    CONSTRAINT SlotFinalizeEpoch FOREIGN KEY (finalizeEpoch) REFERENCES epoch(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO epoch (startDate)
SELECT CURDATE()
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM epoch
);