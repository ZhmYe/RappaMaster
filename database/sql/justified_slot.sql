START TRANSACTION;

SET @current_epoch = (SELECT IFNULL(MAX(id), -1) FROM epoch LOCK IN SHARE MODE);
SET @total_count = ?;

CREATE TEMPORARY TABLE IF NOT EXISTS temp_updated_slots (
    slot_id INT,
    task_id INT,
    slot_size BIGINT
);

INSERT INTO temp_updated_slots (slot_id, task_id, slot_size)
SELECT s.id, t.id, s.finished
FROM slot s
         INNER JOIN task t ON s.taskID = t.id
WHERE s.slotHash IN (?)
  AND s.status = 'committed';

SET @affected_rows = ROW_COUNT();
IF @affected_rows != @total_count THEN
    ROLLBACK;
    SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'slot not exist or not committed',
    MYSQL_ERRNO = 1001;
END IF;

UPDATE slot
SET
    status = 'justified',
    justifiedEpoch = @current_epoch
WHERE
    id IN (SELECT slot_id FROM temp_updated_slots)
  AND status = 'committed';

UPDATE task t
    INNER JOIN (
    SELECT task_id, SUM(slot_size) AS total_size
    FROM temp_updated_slots
    GROUP BY task_id
    ) AS slot_summary ON t.id = slot_summary.task_id
    SET t.size = t.size + slot_summary.total_size;

DROP TEMPORARY TABLE IF EXISTS temp_updated_slots;

COMMIT;