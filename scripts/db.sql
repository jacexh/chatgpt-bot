CREATE TABLE `chat` (
  `id` varchar(32) NOT NULL,
  `counts` tinyint NOT NULL DEFAULT '0',
  `current` text,
  `channel` tinyint NOT NULL,
  `channel_user_id` varchar(45) NOT NULL,
  `version` int NOT NULL DEFAULT '0',
  `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


CREATE TABLE `conversation` (
  `id` int NOT NULL AUTO_INCREMENT,
  `chat_id` varchar(32) DEFAULT NULL,
  `prompt` text NOT NULL,
  `completion` text,
  `channel_message_id` varchar(48) NOT NULL,
  `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `mtime` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
