-- migration: traffic_days
-- up
ALTER TABLE traffic_windows ADD COLUMN view_days TEXT NOT NULL DEFAULT '[]';
ALTER TABLE traffic_windows ADD COLUMN clone_days TEXT NOT NULL DEFAULT '[]';

-- down
