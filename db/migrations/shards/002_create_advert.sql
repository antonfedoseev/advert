CREATE TABLE `advert` (
  `id`          int(11) unsigned NOT NULL AUTO_INCREMENT,
  `owner_id`    int(11) unsigned NOT NULL,  
  `title`       varchar(255) NOT NULL,
  `description` text NOT NULL,
  `сtime`       int(11) unsigned NOT NULL DEFAULT '0',
  `stime`       int(11) unsigned NOT NULL DEFAULT '0',
  `ftime`       int(11) unsigned NOT NULL DEFAULT '0',
  `state`       tinyint(3) unsigned NOT NULL,
  
  PRIMARY KEY   `id`                (`id`),
  KEY           `owner_id`          (`owner_id`),
  KEY           `state`             (`state`),
  FULLTEXT KEY  `title,description` (`title`, `description`)
) CHARACTER SET utf8 COLLATE utf8_general_ci ENGINE=InnoDB;