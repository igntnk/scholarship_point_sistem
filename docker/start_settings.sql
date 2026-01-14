CREATE USER sps_service WITH PASSWORD '314$deTs';
CREATE DATABASE sps_database OWNER sps_service;

GRANT ALL PRIVILEGES ON DATABASE sps_database TO sps_service;
-- Сначала надо покдлючиться к созданной бд sps_database
CREATE SCHEMA sps AUTHORIZATION sps_service;
GRANT ALL PRIVILEGES ON SCHEMA sps TO sps_service;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA sps TO sps_service;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA sps TO sps_service;
ALTER USER sps_service SET search_path TO sps;
ALTER DATABASE sps_database SET search_path TO sps;