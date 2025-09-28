START TRANSACTION;

SET @target_slot_finished = ?;
SET @target_commitment = ?;
SET @target_signature = ?;
SET @target_task_sign_input = ?;
SET @target_slotHash = ?;


SET @current_epoch = (SELECT IFNULL(MAX(id), -1) FROM epoch LOCK IN SHARE MODE);


SET @db_task_sign = (
    SELECT t.sign
    FROM slot s
    INNER JOIN task t ON s.taskID = t.id
    WHERE s.slotHash = @target_slotHash
);

IF @db_task_sign IS NULL OR @db_task_sign != @target_task_sign_input THEN
    ROLLBACK;
    SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'Sign consistency check failed: input sign does not match task''s real sign',
    MYSQL_ERRNO = 1001;
ELSE
UPDATE slot
SET
    commitEpoch = @current_epoch,
    finished = @target_slot_finished,
    commitment = @target_commitment,
    signature = @target_signature,
    status = 'committed'
WHERE
    slotHash = @target_slotHash
    AND status != 'committed';

COMMIT;
END IF;