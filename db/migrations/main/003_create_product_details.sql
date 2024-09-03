CREATE TABLE `product_details` (
  `advert_id`       int(11) unsigned NOT NULL,
  `state`           tinyint(3) unsigned NOT NULL,
  `price`           int(11) unsigned NOT NULL,
  `category`        tinyint(3) unsigned NOT NULL,
  `sub_category_1`  tinyint(3) unsigned NOT NULL,
  `sub_category_2`  tinyint(3) unsigned NOT NULL,
  `sub_category_3`  tinyint(3) unsigned NOT NULL,
  `geolocation`     point NOT NULL,
  `country`         smallint(5) unsigned NOT NULL,
  `area`            smallint(5) unsigned NOT NULL,
  `city`            int(11) unsigned NOT NULL,
  `district`        tinyint(3) unsigned NOT NULL,
    
  PRIMARY KEY   `advert_id`   (`advert_id`),
  KEY           `state`       (`state`),
  KEY           `price`       (`price`),
  KEY           `category`    (`category`, `sub_category_1`, `sub_category_2`, `sub_category_3`),
  SPATIAL INDEX `geolocation` (`geolocation`),
  KEY           `place`       (`country`, `area`, `city`, `district`)
) ENGINE=InnoDB;
