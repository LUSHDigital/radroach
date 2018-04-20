-- The following table contains various common table components such
-- as keys, indexes and relationships and is designed to regression
-- test changes to radroach
CREATE TABLE "weight" (
  "id" bigint(20) unsigned NOT NULL,
  "variant_id" bigint(20) unsigned NOT NULL,
  "version_id" bigint(20) unsigned NOT NULL,
  "size_id" bigint(20) unsigned NOT NULL,
  "master_weight" double DEFAULT NULL,
  "master_weight_unit" enum('g','kg','lb','oz','ml') DEFAULT 'g',
  "gross_weight" double DEFAULT NULL,
  "gross_weight_unit" enum('g','kg','lb','oz','ml') DEFAULT 'g',
  PRIMARY KEY ("id"),
  UNIQUE KEY `variant_id_size_id` (`variant_id`,`size_id`),
  KEY "fk_weight_variant_id" ("variant_id"),
  KEY "fk_weight_version_id" ("version_id"),
  KEY "fk_weight_size_id" ("size_id"),
  CONSTRAINT "fk_weight_size_id" FOREIGN KEY ("size_id") REFERENCES "size" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_weight_variant_id" FOREIGN KEY ("variant_id") REFERENCES "variant" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_weight_version_id" FOREIGN KEY ("version_id") REFERENCES "version" ("id") ON DELETE CASCADE
);