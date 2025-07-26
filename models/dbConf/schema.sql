CREATE TABLE `messages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `recipient_phone` VARCHAR(20) NOT NULL COMMENT 'receipent phone number',
  `content` varchar(200) NOT NULL COMMENT 'message content',
  `status` ENUM('pending', 'failed', 'sent') NOT NULL DEFAULT 'pending',
  `messageId` VARCHAR(100) DEFAULT NULL COMMENT 'message ID from the messaging service',
  `sent_at` datetime DEFAULT NULL COMMENT 'when the message was sent',
  `retry_count` int NOT NULL DEFAULT 0 COMMENT 'number of retry attempts',
  `createdOn` datetime DEFAULT CURRENT_TIMESTAMP COMMENT 'Record creation timestamp',
  `updatedOn` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Record last updated timestamp',
  PRIMARY KEY (`id`),
  KEY `idx_messages_status` (`status`),
  KEY `idx_messages_recipient_phone` (`recipient_phone`)
) ;
