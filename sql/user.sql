CREATE TABLE `user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `network` varchar(50) NOT NULL DEFAULT '',
  `uuid` varchar(100) NOT NULL DEFAULT '',
  `name` varchar(50) DEFAULT NULL,
  `state` varchar(3) NOT NULL,
  `zipcode` int(5) unsigned NOT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `created_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `network` (`network`,`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;