-- create task table to record the tasks submitted by users in frontend
CREATE TABLE IF NOT EXISTS task (
    id INT PRIMARY KEY AUTO_INCREMENT, /* so that we don't need to write an id-allocator */
    sign VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL UNIQUE, /*  so we should limit the length of task */
    expected BIGINT NOT NULL,
    finished BIGINT NOT NULL DEFAULT 0,
    model VARCHAR(50) NOT NULL, /* we not use ENUM or SET since we want to have a dynamic set */
    txHash VARCHAR(64) NOT NULL,  /* transaction hash in blockchain, we must commit the transaction first */
    startDate DATE NOT NULL,
    finishDate DATE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS slot (
    id BIGINT NOT NULL PRIMARY KEY, /* use Snowflake */
    taskID INT NOT NULL,
    nodeID INT NOT NULL, /* commit by which node */
    expected BIGINT NOT NULL,
    finished BIGINT NOT NULL DEFAULT 0,
    scheduleEpoch INT NOT NULL,
    commitEpoch INT NOT NULL, /* commit but not finalized */
    finalizeEpoch INT NOT NULL,
    commitment VARCHAR(64), /* we not constrain not null here, we check in code */
    padding INT NOT NULL DEFAULT 0,
    store ENUM('local', 'ec') DEFAULT 'local',
    CONSTRAINT TaskCommitSlot FOREIGN KEY (taskID) REFERENCES task(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;