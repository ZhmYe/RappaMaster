START TRANSACTION;

SET @target_slot_finished = ?;
SET @target_commitment = ?;
SET @target_signature = ?;
SET @target_slotHash = ?;
SET @target_task_sign = ?;

SET @current_epoch = (SELECT IFNULL(MAX(id), -1) FROM epoch LOCK IN SHARE MODE);

UPDATE slot
SET commitEpoch = @current_epoch,
    finished = @target_slot_finished,
    commitment = @target_commitment,
    signature = @target_signature
WHERE slotHash = @target_slotHash;

SET @slot_updated_rows = ROW_COUNT();
IF @slot_updated_rows = 0 THEN
    ROLLBACK;
    LEAVE;
END IF;

UPDATE task
SET
    finished = finished + @target_slot_finished,
    finishDate = CASE
                     WHEN (finished + @target_slot_finished) > expected
                         THEN NOW()
                     ELSE finishDate
                END,
    done = CASE
               WHEN (finished + @target_slot_finished) > expected
                   THEN TRUE
               ELSE done
            END
WHERE sign = @target_task_sign AND @slot_updated_rows > 0;

COMMIT;