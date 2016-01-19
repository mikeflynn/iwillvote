CREATE TABLE `message` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `network` varchar(50) NOT NULL DEFAULT '',
  `uuid` varchar(100) NOT NULL DEFAULT '',
  `message` text NOT NULL,
  `outgoing` tinyint(1) NOT NULL DEFAULT '1',
  `send_on` timestamp NULL DEFAULT NULL,
  `sent` tinyint(4) DEFAULT '0',
  `created_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `network` (`network`,`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;