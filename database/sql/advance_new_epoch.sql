/* 这里要修改，除了推进意外，还要记录得到上一个epoch的信息，不需要是事务，因为上一个epoch的信息可以不得到，是一个只读 */
INSERT INTO epoch (startDate)
VALUES CURDATE();