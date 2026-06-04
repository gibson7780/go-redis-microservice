-- down
ALTER TABLE stats DROP CONSTRAINT stats_url_id_fkey;
ALTER TABLE stats ADD CONSTRAINT stats_url_id_fkey 
    FOREIGN KEY (url_id) REFERENCES urls(id);