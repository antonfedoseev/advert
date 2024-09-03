CREATE TABLE `product_photo` (
  `id`         tinyint(3) unsigned NOT NULL,
  `advert_id`  int(11) unsigned NOT NULL, 
  `url`        varchar(255) NOT NULL,
  `url_small`  varchar(255) NOT NULL,
  `url_medium` varchar(255) NOT NULL,
  `url_big`    varchar(255) NOT NULL,
  `order`      tinyint(3) unsigned NOT NULL,
  
  PRIMARY KEY `advert_id,id`  (`advert_id`,`id`)
) ENGINE=InnoDB;