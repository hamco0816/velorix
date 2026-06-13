-- 桌面客户端版本发布：管理员上传的更新包元数据。
-- 安装包/blockmap 文件平铺在更新目录根（文件名含版本号），此表只存元数据。
-- 同一 channel 同一时刻只有一条 status='active' 对外提供更新。
CREATE TABLE IF NOT EXISTS desktop_releases (
    id            BIGSERIAL    PRIMARY KEY,
    version       VARCHAR(50)  NOT NULL,
    channel       VARCHAR(20)  NOT NULL DEFAULT 'stable',
    mandatory     BOOLEAN      NOT NULL DEFAULT FALSE,
    notes         TEXT         NOT NULL DEFAULT '',
    setup_file    VARCHAR(255) NOT NULL,
    blockmap_file VARCHAR(255) NOT NULL DEFAULT '',
    latest_yml    TEXT         NOT NULL DEFAULT '',
    file_size     BIGINT       NOT NULL DEFAULT 0,
    status        VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_by    BIGINT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_desktop_releases_version_channel ON desktop_releases (version, channel);
CREATE INDEX IF NOT EXISTS idx_desktop_releases_status ON desktop_releases (status);
CREATE INDEX IF NOT EXISTS idx_desktop_releases_created_at ON desktop_releases (created_at);
