CREATE TABLE `link` (
  `hash` varchar(40) NOT NULL DEFAULT '',
  `user_id` int(11) unsigned DEFAULT NULL,
  `action` varchar(25) NOT NULL DEFAULT '',
  `payload` text,
  `created_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`hash`),
  KEY `user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;