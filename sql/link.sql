CREATE TABLE `link` (
  `hash` varchar(40) NOT NULL DEFAULT '',
  `user_id` int(11) unsigned DEFAULT NULL,
  `action` varchar(25) NOT NULL DEFAULT '',
  `payload` text,
  `created_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expires_in` int(11) DEFAULT NULL,
  `clicks` int(11) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`hash`),
  KEY `user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;