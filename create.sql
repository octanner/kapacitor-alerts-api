do $$
begin

  CREATE TABLE IF NOT EXISTS memory_tasks
  (
    id TEXT NOT NULL UNIQUE,                            -- ID of task (from kapacitor)
    app TEXT NOT NULL,                                  -- Name of app to monitor
    dynotype TEXT NOT NULL,                             -- Dyno to monitor on app
    crit INTEGER NOT NULL,                              -- Threshold for critical alert, in MB
    warn INTEGER NOT NULL,                              -- Threshold for warning alert, in MB
    wind TEXT NOT NULL,                                 -- How far back to retrieve data (e.g. 10m, 30m, 1h)
    every TEXT NOT NULL,                                -- Frequency to check (e.g. 30s, 1m, 10m)
    slack TEXT,                                         -- Slack channel to notify
    post TEXT,                                          -- HTTP endpoint to notify (POST)
    email TEXT,                                         -- Email address to notify
    CONSTRAINT notify_present CHECK (                   -- Got to have a value in either slack, post, or email
      (CASE WHEN slack IS NULL THEN 0 ELSE 1 END) +
      (CASE WHEN post IS NULL THEN 0 ELSE 1 END) +
      (CASE WHEN email IS NULL THEN 0 ELSE 1 END) > 0
    )
  );

  CREATE TABLE IF NOT EXISTS _5xx_tasks 
  (
    app TEXT NOT NULL UNIQUE,                           -- Name of app to monitor
    tolerance TEXT NOT NULL,                            -- How sensitive should checks be? [low, medium, high]
    slack TEXT,                                         -- Slack channel to notify
    post TEXT,                                          -- HTTP endpoint to notify (POST)
    email TEXT,                                         -- Email address to notify
    CONSTRAINT notify_present CHECK (                   -- Got to have a value in either slack, post, or email
      (CASE WHEN slack IS NULL THEN 0 ELSE 1 END) +
      (CASE WHEN post IS NULL THEN 0 ELSE 1 END) +
      (CASE WHEN email IS NULL THEN 0 ELSE 1 END) > 0
    )
  );

  CREATE TABLE IF NOT EXISTS crashed_tasks
  (
    id TEXT NOT NULL,                                   -- ID of task (from kapacitor)
    app TEXT NOT NULL,                                  -- Name of app to monitor
    slack TEXT,                                         -- Slack channel to notify
    post TEXT,                                          -- HTTP endpoint to notify (POST)
    email TEXT,                                         -- Email address to notify
    CONSTRAINT notify_present CHECK (                   -- Got to have a value in either slack, post, or email
      (CASE WHEN slack IS NULL THEN 0 ELSE 1 END) +
      (CASE WHEN post IS NULL THEN 0 ELSE 1 END) +
      (CASE WHEN email IS NULL THEN 0 ELSE 1 END) > 0
    )
  );

  CREATE TABLE IF NOT EXISTS release_tasks
  (
    id TEXT NOT NULL,                                   -- ID of task (from kapacitor)
    app TEXT NOT NULL,                                  -- Name of app to monitor
    slack TEXT,                                         -- Slack channel to notify
    post TEXT,                                          -- HTTP endpoint to notify (POST)
    email TEXT,                                         -- Email address to notify
    CONSTRAINT notify_present CHECK (                   -- Got to have a value in either slack, post, or email
      (CASE WHEN slack IS NULL THEN 0 ELSE 1 END) +
      (CASE WHEN post IS NULL THEN 0 ELSE 1 END) +
      (CASE WHEN email IS NULL THEN 0 ELSE 1 END) > 0
    )
  );

end
$$;
