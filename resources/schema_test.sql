DROP TABLE IF EXISTS instances CASCADE;
DROP TABLE IF EXISTS works CASCADE;
DROP TABLE IF EXISTS series CASCADE;
DROP TABLE IF EXISTS listings CASCADE;
DROP TABLE IF EXISTS groups CASCADE;
DROP TABLE IF EXISTS contents CASCADE;
DROP TABLE IF EXISTS texts CASCADE;
DROP TABLE IF EXISTS media CASCADE;
DROP TABLE IF EXISTS sources CASCADE;
DROP TABLE IF EXISTS captions CASCADE;
DROP TABLE IF EXISTS fileservers CASCADE;
DROP TABLE IF EXISTS filenodes CASCADE;
DROP TABLE IF EXISTS files CASCADE;
DROP TABLE IF EXISTS filenames CASCADE;

DROP TYPE IF EXISTS l_type CASCADE;
DROP TYPE IF EXISTS c_type CASCADE;

CREATE TYPE l_type AS ENUM ('series', 'work');
CREATE TYPE c_type AS ENUM ('group', 'text', 'media');

CREATE TABLE instances (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    scheme VARCHAR(8) NOT NULL,
    host VARCHAR(4096) NOT NULL,
    port INTEGER NOT NULL,
    title VARCHAR(4096) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE (host, port)
);

CREATE TABLE listings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    instance_id BIGINT,
    parent_series_id BIGINT,
    position NUMERIC NOT NULL,
    listing_type L_TYPE NOT NULL,
    CHECK ((instance_id IS NULL) <> (parent_series_id IS NULL))
);

CREATE TABLE works (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    listing_id BIGINT UNIQUE NOT NULL,
    title VARCHAR(4096) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE series (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    listing_id BIGINT UNIQUE NOT NULL,
    title VARCHAR(4096) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE contents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    work_id BIGINT,
    parent_group_id BIGINT,
    position NUMERIC NOT NULL,
    content_type C_TYPE NOT NULL,
    CHECK ((work_id IS NULL) <> (parent_group_id IS NULL))
);

CREATE TABLE groups (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    content_id BIGINT UNIQUE NOT NULL,
    swap_direction BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE TABLE texts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    content_id BIGINT UNIQUE NOT NULL,
    content VARCHAR(4194304)
);

CREATE TABLE media (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    content_id BIGINT UNIQUE NOT NULL,
    file_hash VARCHAR(128),
    caption VARCHAR(4194304)
);

CREATE TABLE fileservers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    scheme VARCHAR(8) NOT NULL,
    host VARCHAR(4096) NOT NULL,
    port INTEGER NOT NULL,
    disk_path VARCHAR(4096) NOT NULL,
    url_path VARCHAR(4096) NOT NULL,
    UNIQUE (host, port)
);

CREATE TABLE files (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    size BIGINT NOT NULL,
    hash VARCHAR(128) UNIQUE NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    CHECK (size >= 0)
);

CREATE TABLE filenodes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    file_id BIGINT NOT NULL,
    fileserver_id BIGINT NOT NULL,
    path VARCHAR(4096) NOT NULL,
    UNIQUE (fileserver_id, path)
);

CREATE INDEX filenodes_file_id_idx ON filenodes (file_id);
CREATE INDEX filenodes_fileserver_id_idx ON filenodes (fileserver_id);
CREATE INDEX media_file_hash_idx ON media (file_hash);

CREATE UNIQUE INDEX listings_instance_position_unique
ON listings (instance_id, position)
WHERE parent_series_id IS NULL;

CREATE UNIQUE INDEX listings_parent_series_position_unique
ON listings (parent_series_id, position)
WHERE parent_series_id IS NOT NULL;

CREATE UNIQUE INDEX contents_work_position_unique
ON contents (work_id, position)
WHERE parent_group_id IS NULL;

CREATE UNIQUE INDEX contents_parent_group_position_unique
ON contents (parent_group_id, position)
WHERE parent_group_id IS NOT NULL;

ALTER TABLE listings
ADD CONSTRAINT listings_instance_fk
FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE;

ALTER TABLE listings
ADD CONSTRAINT listings_parent_series_fk
FOREIGN KEY (parent_series_id) REFERENCES series(id) ON DELETE CASCADE;

ALTER TABLE series
ADD CONSTRAINT series_listing_fk
FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE;

ALTER TABLE works
ADD CONSTRAINT works_listing_fk
FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE;

ALTER TABLE contents
ADD CONSTRAINT contents_work_fk
FOREIGN KEY (work_id) REFERENCES works(id) ON DELETE CASCADE;

ALTER TABLE contents
ADD CONSTRAINT contents_parent_group_fk
FOREIGN KEY (parent_group_id) REFERENCES groups(id) ON DELETE CASCADE;

ALTER TABLE groups
ADD CONSTRAINT groups_content_fk
FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE;

ALTER TABLE texts
ADD CONSTRAINT texts_content_fk
FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE;

ALTER TABLE media
ADD CONSTRAINT media_content_fk
FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE;

ALTER TABLE media
ADD CONSTRAINT media_file_fk
FOREIGN KEY (file_hash) REFERENCES files(hash) ON DELETE RESTRICT;

ALTER TABLE filenodes
ADD CONSTRAINT filenodes_file_fk
FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE;

ALTER TABLE filenodes
ADD CONSTRAINT filenodes_fileserver_fk
FOREIGN KEY (fileserver_id) REFERENCES fileservers(id) ON DELETE CASCADE;

-- ===== mock data =====

-- The local file server, matching the static file handlers registered in main.
INSERT INTO fileservers (scheme, host, port, disk_path, url_path)
VALUES ('http', 'localhost', 8080, './resources/media', '/media/'); -- fileserver.id 1

-- Sizes and digests of the seeded files under the objects directory, which is
-- what an upload of those same bytes would compute.
INSERT INTO files (size, hash, mime_type)
VALUES (3215511, '2ccae0bc65d10ced9dd9d2404a9f39fd4505180f447b44461bb7ebcb363f4aa1', 'video/mp4'); -- file.id 1

INSERT INTO files (size, hash, mime_type)
VALUES (35911, 'ef7b88411629d1983f2c5b4b78351e6f708f2b2105de30d37ebc917568d504d4', 'image/png'); -- file.id 2

INSERT INTO files (size, hash, mime_type)
VALUES (3015647, '494664c7bde01ba414e801db4538063991e57cf046e63c51511ae6f86b1a01fc', 'audio/ogg'); -- file.id 3

-- A path is relative to the objects directory of the Fileserver, and is the
-- digest of the file sharded over two directories, under the extension its
-- MIME type is stored as.
INSERT INTO filenodes (file_id, fileserver_id, path)
VALUES (1, 1, '2c/ca/2ccae0bc65d10ced9dd9d2404a9f39fd4505180f447b44461bb7ebcb363f4aa1.mp4'); -- filenode.id 1

INSERT INTO filenodes (file_id, fileserver_id, path)
VALUES (2, 1, 'ef/7b/ef7b88411629d1983f2c5b4b78351e6f708f2b2105de30d37ebc917568d504d4.png'); -- filenode.id 2

INSERT INTO filenodes (file_id, fileserver_id, path)
VALUES (3, 1, '49/46/494664c7bde01ba414e801db4538063991e57cf046e63c51511ae6f86b1a01fc.oga'); -- filenode.id 3

-- The address the Instance is reached on, which is the one it redirects to:
-- the Nginx port, not the port this program listens on.
INSERT INTO instances (scheme, host, port, title)
VALUES ('http', 'localhost', 80, 'WorksKeeper'); -- instance.id 1

-- Top-level Series belong directly to the Instance through their Listings.
INSERT INTO listings (instance_id, parent_series_id, position, listing_type)
VALUES (1, NULL, 1.0, 'series'); -- listing.id 1: Series #2

INSERT INTO series (listing_id, title, created_at)
VALUES (1, 'Series #2', NOW()); -- series.id 1

INSERT INTO listings (instance_id, parent_series_id, position, listing_type)
VALUES (1, NULL, 2.0, 'series'); -- listing.id 2: Series #3

INSERT INTO series (listing_id, title, created_at)
VALUES (2, 'Series #3', NOW()); -- series.id 2

-- Works
INSERT INTO listings (instance_id, parent_series_id, position, listing_type)
VALUES (NULL, 1, 1.0, 'work'); -- listing.id 3: Work #1 under Series #2

INSERT INTO listings (instance_id, parent_series_id, position, listing_type)
VALUES (1, NULL, 3.0, 'work'); -- listing.id 4: Work #2 at the top level

INSERT INTO works (listing_id, title)
VALUES (3, 'Work #1 (nested)'); -- work.id 1

INSERT INTO works (listing_id, title)
VALUES (4, 'Work #2'); -- work.id 2

-- Work #1 contents: a top-level group + a top-level text

INSERT INTO contents (work_id, parent_group_id, position, content_type)
VALUES (1, NULL, 1.0, 'group'); -- content.id 1

INSERT INTO groups (content_id) VALUES (1); -- group.id 1

INSERT INTO contents (work_id, parent_group_id, position, content_type)
VALUES (1, NULL, 2.0, 'text'); -- content.id 2

INSERT INTO texts (content_id, content)
VALUES (2, 'Text #2');

-- Inside Group #1: a text, then a video with a caption

INSERT INTO contents (work_id, parent_group_id, position, content_type)
VALUES (NULL, 1, 1.0, 'text'); -- content.id 3

INSERT INTO texts (content_id, content)
VALUES (3, 'Text #1');

INSERT INTO contents (work_id, parent_group_id, position, content_type)
VALUES (NULL, 1, 2.0, 'media'); -- content.id 4

INSERT INTO media (content_id, file_hash, caption)
VALUES (4, '2ccae0bc65d10ced9dd9d2404a9f39fd4505180f447b44461bb7ebcb363f4aa1', 'Caption #1'); -- media.id 1

-- Work #2 contents: a single top-level text

INSERT INTO contents (work_id, parent_group_id, position, content_type)
VALUES (2, NULL, 1.0, 'text'); -- content.id 5

INSERT INTO texts (content_id, content)
VALUES (5, 'Text #3');

-- Series #3 contents: a Work exercising the image and sound Media branches

INSERT INTO listings (instance_id, parent_series_id, position, listing_type)
VALUES (NULL, 2, 1.0, 'work'); -- listing.id 5: Work #3 under Series #3

INSERT INTO works (listing_id, title)
VALUES (5, 'Work #3 (in Series #3)'); -- work.id 3

INSERT INTO contents (work_id, parent_group_id, position, content_type)
VALUES (3, NULL, 1.0, 'media'); -- content.id 6

INSERT INTO media (content_id, file_hash, caption)
VALUES (6, 'ef7b88411629d1983f2c5b4b78351e6f708f2b2105de30d37ebc917568d504d4', 'Caption #2 (on an image)'); -- media.id 2

INSERT INTO contents (work_id, parent_group_id, position, content_type)
VALUES (3, NULL, 2.0, 'media'); -- content.id 7

INSERT INTO media (content_id, file_hash)
VALUES (7, '494664c7bde01ba414e801db4538063991e57cf046e63c51511ae6f86b1a01fc'); -- media.id 3