CREATE TABLE `user_message` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `message_id` int(11) unsigned NOT NULL,
  `network` varchar(50) NOT NULL DEFAULT '',
  `uuid` varchar(100) NOT NULL DEFAULT '',
  `params` varchar(200) NOT NULL DEFAULT '',
  `send_on` timestamp NULL DEFAULT NULL,
  `sent` tinyint(1) NOT NULL DEFAULT '0',
  `created_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `message_id` (`message_id`),
  KEY `network` (`network`,`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;