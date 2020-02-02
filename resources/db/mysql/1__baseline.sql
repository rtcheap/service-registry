-- +migrate Up
CREATE TABLE `service` (
  `id` VARCHAR(50) NOT NULL,
  `application` VARCHAR(100) NOT NULL,
  `location` VARCHAR(100) NOT NULL,
  `port` INT NOT NULL,
  `status` VARCHAR(20) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE(`location`, `port`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
CREATE INDEX `idx_service_application` ON `service`(`application`);
-- +migrate Down
DROP INDEX `idx_service_application` ON `service`;
DROP TABLE IF EXISTS `service`;