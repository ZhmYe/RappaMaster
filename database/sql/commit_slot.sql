START TRANSACTION;

SET @target_slot_finished = ?;
SET @target_commitment = ?;
SET @target_signature = ?;
SET @target_slotHash = ?;

SET @current_epoch = (SELECT IFNULL(MAX(id), -1) FROM epoch LOCK IN SHARE MODE);

UPDATE slot
SET commitEpoch = @current_epoch,
    finished = @target_slot_finished,
    commitment = @target_commitment,
    signature = @target_signature,
    status = 'committed'
WHERE slotHash = @target_slotHash;

SET @slot_updated_rows = ROW_COUNT();
IF @slot_updated_rows = 0 THEN
    ROLLBACK;
    LEAVE;
END IF;

COMMIT;