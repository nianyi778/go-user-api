-- ============================================================
-- Go User API 数据库初始化脚本
-- ============================================================
-- 本脚本用于初始化 MySQL 数据库
-- 使用方法: mysql -u root -p < scripts/init_db.sql
--
-- 注意: 如果使用 SQLite，不需要运行此脚本，
--       应用程序会自动创建数据库文件和表结构
-- ============================================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `go_user_api`
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE `go_user_api`;

-- ============================================================
-- 用户表
-- ============================================================
CREATE TABLE IF NOT EXISTS `users` (
    -- 主键，使用 UUID
    `id` VARCHAR(36) NOT NULL,

    -- 用户基本信息
    `username` VARCHAR(50) NOT NULL COMMENT '用户名，唯一',
    `email` VARCHAR(100) NOT NULL COMMENT '邮箱，唯一',
    `password` VARCHAR(255) NOT NULL COMMENT '密码哈希值',
    `nickname` VARCHAR(50) DEFAULT '' COMMENT '昵称',
    `avatar` VARCHAR(255) DEFAULT '' COMMENT '头像 URL',
    `phone` VARCHAR(20) DEFAULT '' COMMENT '手机号',
    `bio` VARCHAR(500) DEFAULT '' COMMENT '个人简介',
    `gender` TINYINT DEFAULT 0 COMMENT '性别: 0-未知, 1-男, 2-女',
    `birthday` DATE DEFAULT NULL COMMENT '生日',

    -- 用户状态和角色
    `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-正常, 2-未激活',
    `role` VARCHAR(20) DEFAULT 'user' COMMENT '角色: user, admin',

    -- 登录信息
    `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
    `last_login_ip` VARCHAR(45) DEFAULT '' COMMENT '最后登录 IP',

    -- 时间戳
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间（软删除）',

    -- 主键
    PRIMARY KEY (`id`),

    -- 唯一索引
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_email` (`email`),

    -- 普通索引
    KEY `idx_phone` (`phone`),
    KEY `idx_status` (`status`),
    KEY `idx_role` (`role`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_created_at` (`created_at`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================================
-- 插入管理员账号（可选）
-- ============================================================
-- 密码: admin123 (bcrypt 加密)
-- 注意: 生产环境请修改密码！
INSERT INTO `users` (
    `id`,
    `username`,
    `email`,
    `password`,
    `nickname`,
    `status`,
    `role`,
    `created_at`,
    `updated_at`
) VALUES (
    UUID(),
    'admin',
    'admin@example.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeliQe29x67sQQ2C.VLz3U5gKxEfWa7Ym', -- admin123
    'Administrator',
    1,
    'admin',
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- ============================================================
-- 创建只读用户（可选，用于报表等只读场景）
-- ============================================================
-- CREATE USER IF NOT EXISTS 'go_user_api_readonly'@'%' IDENTIFIED BY 'readonly_password';
-- GRANT SELECT ON go_user_api.* TO 'go_user_api_readonly'@'%';

-- ============================================================
-- 创建应用程序用户（推荐）
-- ============================================================
-- CREATE USER IF NOT EXISTS 'go_user_api_app'@'%' IDENTIFIED BY 'your_secure_password';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON go_user_api.* TO 'go_user_api_app'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- ============================================================
-- 查看表结构
-- ============================================================
-- DESCRIBE users;

-- ============================================================
-- 查看索引
-- ============================================================
-- SHOW INDEX FROM users;

-- 完成
SELECT 'Database initialization completed!' AS message;
