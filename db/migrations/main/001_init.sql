CREATE TABLE `schema_migrations` (
  `id`        int(11) unsigned NOT NULL AUTO_INCREMENT,
  `migration` varchar(255) NOT NULL,
  `сtime`     int(11) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY `id` (`id`),
  KEY `migration` (`migration`)
) ENGINE=InnoDB;