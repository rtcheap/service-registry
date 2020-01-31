-- +migrate Up
CREATE TABLE `service` (
  `id` VARCHAR(50) NOT NULL,
  `application` VARCHAR(100) NOT NULL,
  `location` VARCHAR(100) NOT NULL,
  `port` INTEGER NOT NULL,
  `status` VARCHAR(20) NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE(`application`, `location`, `port`)
);
CREATE INDEX `idx_service_application` ON `service`(`application`);
-- +migrate Down
DROP INDEX IF EXISTS `idx_service_application`;
DROP TABLE IF EXISTS `service`;