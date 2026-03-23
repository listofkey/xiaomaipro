/*
 Navicat Premium Dump SQL

 Source Server         : 127.0.0.1
 Source Server Type    : PostgreSQL
 Source Server Version : 160013 (160013)
 Source Host           : localhost:5432
 Source Catalog        : xiaomaipro
 Source Schema         : public

 Target Server Type    : PostgreSQL
 Target Server Version : 160013 (160013)
 File Encoding         : 65001

 Date: 24/03/2026 01:23:22
*/


-- ----------------------------
-- Sequence structure for category_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."category_id_seq";
CREATE SEQUENCE "public"."category_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for category_id_seq1
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."category_id_seq1";
CREATE SEQUENCE "public"."category_id_seq1" 
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;

-- ----------------------------
-- Table structure for address
-- ----------------------------
DROP TABLE IF EXISTS "public"."address";
CREATE TABLE "public"."address" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "receiver_name" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "receiver_phone" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "province" varchar(32) COLLATE "pg_catalog"."default" NOT NULL,
  "city" varchar(32) COLLATE "pg_catalog"."default" NOT NULL,
  "district" varchar(32) COLLATE "pg_catalog"."default" NOT NULL,
  "detail" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "is_default" int2 NOT NULL DEFAULT 0,
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."address" IS 'Shipping address table';

-- ----------------------------
-- Records of address
-- ----------------------------
INSERT INTO "public"."address" VALUES (2035038423353401344, 2035026620841992192, '123', '12312341234', '上海', '上海', '123', '上海市杨浦区邯郸路55号', 0, '2026-03-21 00:59:03.971', '2026-03-21 01:54:30.305');
INSERT INTO "public"."address" VALUES (2035038354281603072, 2035026620841992192, 'lihua', '12312341234', '上海', '上海', '12', '上海市杨浦区邯郸路55号', 1, '2026-03-21 00:58:47.503', '2026-03-21 01:54:30.305');
INSERT INTO "public"."address" VALUES (120000000000000001, 100000000000000001, '陈曦', '13800010001', '北京市', '北京市', '朝阳区', '望京街道阜通东大街6号院3号楼1602', 1, '2025-12-16 09:40:00', '2026-03-20 09:40:00');
INSERT INTO "public"."address" VALUES (120000000000000002, 100000000000000002, '卢露', '13800010002', '上海市', '上海市', '浦东新区', '芳甸路1188弄6号1201', 1, '2025-12-19 10:40:00', '2026-03-20 10:40:00');
INSERT INTO "public"."address" VALUES (120000000000000003, 100000000000000003, '莫莫', '13800010003', '广东省', '深圳市', '南山区', '后海大道88号科苑公馆2栋901', 1, '2026-01-04 11:50:00', '2026-03-20 11:50:00');
INSERT INTO "public"."address" VALUES (120000000000000004, 100000000000000004, '赵凯文', '13800010004', '浙江省', '杭州市', '上城区', '钱江路58号新城时代广场A座2008', 1, '2026-01-11 14:50:00', '2026-03-20 14:50:00');
INSERT INTO "public"."address" VALUES (120000000000000005, 100000000000000005, '李文雯', '13800010005', '四川省', '成都市', '高新区', '天府三街199号环球中心西区1306', 1, '2026-01-17 15:50:00', '2026-03-20 15:50:00');
INSERT INTO "public"."address" VALUES (120000000000000006, 100000000000000006, '孙娜娜', '13800010006', '江苏省', '南京市', '建邺区', '江东中路188号金鹰世界B座2605', 1, '2026-02-03 16:50:00', '2026-03-20 16:50:00');

-- ----------------------------
-- Table structure for admin_log
-- ----------------------------
DROP TABLE IF EXISTS "public"."admin_log";
CREATE TABLE "public"."admin_log" (
  "id" int8 NOT NULL,
  "admin_id" int8 NOT NULL,
  "action" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "target_type" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "target_id" int8,
  "detail" text COLLATE "pg_catalog"."default",
  "ip" varchar(45) COLLATE "pg_catalog"."default",
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."admin_log" IS 'Admin audit log table';

-- ----------------------------
-- Records of admin_log
-- ----------------------------
INSERT INTO "public"."admin_log" VALUES (210000000000000001, 900000000000000002, 'create_event', 'event', 140000000000000001, '{"title":"周杰伦 2026 嘉年华世界巡回演唱会 - 北京站","status":1}', '10.10.0.11', '2026-02-20 10:05:00');
INSERT INTO "public"."admin_log" VALUES (210000000000000002, 900000000000000003, 'publish_event', 'event', 140000000000000010, '{"title":"阿那亚戏剧节特别场《海边的卡夫卡》 - 北京站","status":1}', '10.10.0.12', '2026-03-14 16:30:00');
INSERT INTO "public"."admin_log" VALUES (210000000000000003, 900000000000000002, 'off_shelf_event', 'event', 140000000000000011, '{"reason":"档期调整","status":2}', '10.10.0.13', '2026-03-20 12:30:00');
INSERT INTO "public"."admin_log" VALUES (210000000000000004, 900000000000000004, 'audit_refund', 'refund', 190000000000000001, '{"result":"approved","refundNo":"RFD202603200001"}', '10.10.0.14', '2026-03-20 15:40:00');
INSERT INTO "public"."admin_log" VALUES (210000000000000005, 900000000000000003, 'batch_notify', 'event', 140000000000000005, '{"scene":"hot_sale_push","userCount":128}', '10.10.0.15', '2026-03-20 19:58:00');

-- ----------------------------
-- Table structure for admin_user
-- ----------------------------
DROP TABLE IF EXISTS "public"."admin_user";
CREATE TABLE "public"."admin_user" (
  "id" int8 NOT NULL,
  "username" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "password_hash" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "real_name" varchar(64) COLLATE "pg_catalog"."default",
  "phone" varchar(20) COLLATE "pg_catalog"."default",
  "role" int2 NOT NULL,
  "status" int2 NOT NULL DEFAULT 1,
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."admin_user" IS 'Admin user table';

-- ----------------------------
-- Records of admin_user
-- ----------------------------
INSERT INTO "public"."admin_user" VALUES (900000000000000001, 'superadmin', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '系统管理员', '13900000001', 1, 1, '2025-12-01 09:00:00', '2026-03-20 09:00:00');
INSERT INTO "public"."admin_user" VALUES (900000000000000002, 'ops_liu', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '刘运营', '13900000002', 2, 1, '2025-12-05 10:00:00', '2026-03-20 10:00:00');
INSERT INTO "public"."admin_user" VALUES (900000000000000003, 'ops_wang', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '王运营', '13900000003', 2, 1, '2025-12-08 11:00:00', '2026-03-20 11:00:00');
INSERT INTO "public"."admin_user" VALUES (900000000000000004, 'cs_lin', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '林客服', '13900000004', 3, 1, '2025-12-10 14:00:00', '2026-03-20 14:00:00');

-- ----------------------------
-- Table structure for category
-- ----------------------------
DROP TABLE IF EXISTS "public"."category";
CREATE TABLE "public"."category" (
  "id" int4 NOT NULL GENERATED BY DEFAULT AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1
),
  "name" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "icon" varchar(255) COLLATE "pg_catalog"."default",
  "sort_order" int4 NOT NULL DEFAULT 0,
  "status" int2 NOT NULL DEFAULT 1
)
;
COMMENT ON TABLE "public"."category" IS 'Event category table';

-- ----------------------------
-- Records of category
-- ----------------------------
INSERT INTO "public"."category" VALUES (1, '演唱会', 'https://cdn.example.com/icons/concert.png', 10, 1);
INSERT INTO "public"."category" VALUES (2, '话剧歌剧', 'https://cdn.example.com/icons/drama.png', 20, 1);
INSERT INTO "public"."category" VALUES (3, '音乐节', 'https://cdn.example.com/icons/festival.png', 30, 1);
INSERT INTO "public"."category" VALUES (4, '体育赛事', 'https://cdn.example.com/icons/sports.png', 40, 1);
INSERT INTO "public"."category" VALUES (5, '展览休闲', 'https://cdn.example.com/icons/exhibit.png', 50, 1);
INSERT INTO "public"."category" VALUES (6, '亲子演出', 'https://cdn.example.com/icons/family.png', 60, 1);

-- ----------------------------
-- Table structure for city
-- ----------------------------
DROP TABLE IF EXISTS "public"."city";
CREATE TABLE "public"."city" (
  "id" int8 NOT NULL,
  "name" varchar(255) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Records of city
-- ----------------------------
INSERT INTO "public"."city" VALUES (1, '北京');
INSERT INTO "public"."city" VALUES (2, '上海');
INSERT INTO "public"."city" VALUES (3, '广州');
INSERT INTO "public"."city" VALUES (4, '杭州');
INSERT INTO "public"."city" VALUES (5, '成都');
INSERT INTO "public"."city" VALUES (6, '南京');

-- ----------------------------
-- Table structure for event
-- ----------------------------
DROP TABLE IF EXISTS "public"."event";
CREATE TABLE "public"."event" (
  "id" int8 NOT NULL,
  "title" varchar(200) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "poster_url" varchar(255) COLLATE "pg_catalog"."default",
  "category_id" int4 NOT NULL,
  "venue_id" int8 NOT NULL,
  "city_id" int8 NOT NULL,
  "artist" varchar(128) COLLATE "pg_catalog"."default",
  "event_start_time" timestamp(3) NOT NULL,
  "event_end_time" timestamp(3) NOT NULL,
  "sale_start_time" timestamp(3) NOT NULL,
  "sale_end_time" timestamp(3) NOT NULL,
  "status" int2 NOT NULL DEFAULT 0,
  "purchase_limit" int4 NOT NULL DEFAULT 1,
  "need_real_name" int2 NOT NULL DEFAULT 0,
  "ticket_type" int2 NOT NULL DEFAULT 1,
  "created_by" int8 NOT NULL,
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamp(3)
)
;
COMMENT ON TABLE "public"."event" IS 'Event master table';

-- ----------------------------
-- Records of event
-- ----------------------------
INSERT INTO "public"."event" VALUES (140000000000000001, 'ONEREPUBLIC "From Asia, With Love" 2026北京站', '延续嘉年华主题舞台设计，北京站设置四面台与限定曲目返场，适合核心粉丝和家庭观演。', 'https://img.alicdn.com/bao/uploaded/i2/2251059038/O1CN01E3LfNG2GdSlslh9hu_!!4611686018427383646-0-item_pic.jpg', 1, 130000000000000001, 1, '周杰伦', '2026-04-18 19:30:00', '2026-04-18 22:30:00', '2026-03-01 11:00:00', '2026-04-18 20:00:00', 1, 4, 1, 1, 900000000000000002, '2026-02-20 10:00:00', '2026-03-24 00:43:47.078', NULL);
INSERT INTO "public"."event" VALUES (140000000000000012, '上海歌舞团《永不消逝的电波》-泱泱国风·舞动经典', '含赛前热身、完赛音乐派对和品牌体验区，活动已于 2025 年 11 月结束。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i3/2251059038/O1CN01sr0Ih02GdSlMsWQym_!!2251059038.png', 2, 130000000000000005, 1, '广州市体育局', '2025-11-09 07:30:00', '2025-11-09 12:30:00', '2025-09-15 10:00:00', '2025-11-09 08:00:00', 3, 2, 1, 1, 900000000000000002, '2025-08-01 09:00:00', '2026-03-24 00:54:47.147', NULL);
INSERT INTO "public"."event" VALUES (140000000000000013, '《2026檀谷开山节》(北京檀谷)', '仍处于排期确认阶段，暂未正式上架。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i3/2251059038/O1CN01XO3KRm2GdSlIDrnPb_!!2251059038.png', 2, 130000000000000009, 2, '星空魔术团', '2026-07-02 19:00:00', '2026-07-02 20:40:00', '2026-04-01 10:00:00', '2026-07-02 18:00:00', 0, 6, 0, 2, 900000000000000003, '2026-03-16 11:00:00', '2026-03-24 00:54:55.404', NULL);
INSERT INTO "public"."event" VALUES (140000000000000002, '声名远扬 2026·华语榜中榜金曲演唱会', '上海站为室内四面台版本，加入全新编曲和纪念环节，适合情侣和朋友结伴观演。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i4/2251059038/O1CN01lj5hbe2GdSlmsQoyy_!!2251059038.png', 1, 130000000000000002, 2, '林俊杰', '2026-04-25 19:30:00', '2026-04-25 22:15:00', '2026-03-05 10:00:00', '2026-04-25 20:00:00', 1, 4, 1, 1, 900000000000000003, '2026-02-25 11:00:00', '2026-03-24 00:46:28.54', NULL);
INSERT INTO "public"."event" VALUES (140000000000000005, '2026光良“不会分离·今晚我不孤独3.0⁺”演唱会-北京站', '高关注度焦点赛事，现场配套灯光秀和球迷互动区，实名电子票入场。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i2/2251059038/O1CN01AteXtk2GdSlWdZb8z_!!2251059038.jpg', 3, 130000000000000005, 3, '广州龙狮 vs 辽宁本钢', '2026-05-16 19:35:00', '2026-05-16 22:00:00', '2026-03-12 12:00:00', '2026-05-16 19:00:00', 1, 2, 1, 1, 900000000000000002, '2026-03-03 11:00:00', '2026-03-24 00:52:49.465', NULL);
INSERT INTO "public"."event" VALUES (140000000000000008, 'ONEREPUBLIC "From Asia, With Love" 2026北京站', '广州站加开场次，经典曲目占比高，适合家庭和资深歌迷。', 'https://img.alicdn.com/bao/uploaded/i2/2251059038/O1CN01E3LfNG2GdSlslh9hu_!!4611686018427383646-0-item_pic.jpg', 1, 130000000000000005, 3, '张学友', '2026-06-20 19:30:00', '2026-06-20 22:20:00', '2026-03-18 10:00:00', '2026-06-20 20:00:00', 1, 4, 1, 1, 900000000000000003, '2026-03-09 12:00:00', '2026-03-24 00:48:29.422', NULL);
INSERT INTO "public"."event" VALUES (140000000000000003, '黄诗扶「入梦」音乐幕剧', '覆盖摇滚、流行和独立音乐阵容，含创意市集、露营区和餐饮区，适合周末短途出游。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i1/2251059038/O1CN01viMDui2GdSlVtzfvf_!!2251059038.png', 1, 130000000000000003, 2, '群星', '2026-05-02 13:00:00', '2026-05-03 22:00:00', '2026-03-08 10:00:00', '2026-05-03 18:00:00', 1, 6, 0, 1, 900000000000000002, '2026-02-28 14:00:00', '2026-03-24 00:48:44.363', NULL);
INSERT INTO "public"."event" VALUES (140000000000000006, '2026 DLC梦想日・北京站', '长周期沉浸式数字展，适合情侣打卡和亲子周末活动，分工作日票与周末票。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i3/2251059038/O1CN011CsUm12GdSlpJLLyP_!!2251059038.jpg', 4, 130000000000000007, 1, '法国印象派主题展', '2026-05-01 10:00:00', '2026-06-30 21:30:00', '2026-03-15 10:00:00', '2026-06-30 18:00:00', 1, 8, 0, 1, 900000000000000003, '2026-03-05 13:00:00', '2026-03-24 00:52:51.586', NULL);
INSERT INTO "public"."event" VALUES (140000000000000007, '《2026檀谷开山节》(北京檀谷)', '90分钟沉浸式亲子音乐剧，适合4岁以上儿童，支持家庭套票。', 'https://img.alicdn.com/bao/uploaded/i4/2251059038/O1CN01Q4OaO42GdSlpqDR8N_!!4611686018427383646-2-item_pic.png', 5, 130000000000000006, 1, '蓝精灵剧团', '2026-04-12 15:00:00', '2026-04-12 17:00:00', '2026-03-10 10:30:00', '2026-04-12 14:00:00', 1, 6, 0, 2, 900000000000000002, '2026-03-07 09:00:00', '2026-03-24 00:52:53.349', NULL);
INSERT INTO "public"."event" VALUES (140000000000000004, '孟慧圆「一场30000天的游戏」2026北京站', '经典舞剧巡演杭州返场，适合家庭观众和舞蹈爱好者，支持纸质纪念票寄送。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i3/2251059038/O1CN019A7cJ22GdSleUIx6U_!!2251059038.jpg', 6, 130000000000000004, 2, '中国东方演艺集团', '2026-05-09 19:30:00', '2026-05-09 21:45:00', '2026-03-10 10:00:00', '2026-05-09 18:30:00', 1, 6, 0, 2, 900000000000000003, '2026-03-01 09:30:00', '2026-03-24 00:53:01.282', NULL);
INSERT INTO "public"."event" VALUES (140000000000000011, '五月天 诺亚方舟十周年特别版 - 南京站', '原计划暑期档开演，因档期调整暂时下架，已支付订单支持无损退款。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i2/2251059038/O1CN01W73kTQ2GdSlm8Fpzs_!!2251059038.jpg', 2, 130000000000000008, 3, '五月天', '2026-07-18 19:30:00', '2026-07-18 22:30:00', '2026-03-12 10:00:00', '2026-07-18 20:00:00', 2, 4, 1, 1, 900000000000000002, '2026-03-15 09:00:00', '2026-03-24 00:55:03.724', NULL);
INSERT INTO "public"."event" VALUES (140000000000000009, '【明星场】开心麻花惊怂爆笑贺岁大戏《出马》', '退役球星友谊赛与球迷互动专场，适合体育粉丝和亲子观赛。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i4/2251059038/O1CN01yqyKs82GdSlrqCLf1_!!2251059038.jpg', 2, 130000000000000002, 2, '世界传奇球星联队', '2026-06-06 19:00:00', '2026-06-06 21:30:00', '2026-03-20 10:00:00', '2026-06-06 18:30:00', 1, 4, 1, 1, 900000000000000002, '2026-03-12 10:00:00', '2026-03-24 00:54:10.637', NULL);
INSERT INTO "public"."event" VALUES (140000000000000010, '话剧《第七天》  余华原著×孟京辉导演 陈明昊 黄湘丽 李庚希 领衔主演', '限量小剧场版本，强调沉浸叙事与舞台设计，适合戏剧重度用户。', 'https://img.alicdn.com/bao/uploaded/https://img.alicdn.com/imgextra/i1/2251059038/O1CN01u31bn62GdSlpS1sUP_!!2251059038.png', 2, 130000000000000010, 1, '阿那亚戏剧节', '2026-06-12 19:30:00', '2026-06-12 21:50:00', '2026-03-22 10:00:00', '2026-06-12 18:30:00', 1, 4, 0, 2, 900000000000000003, '2026-03-14 16:00:00', '2026-03-24 00:54:33.878', NULL);

-- ----------------------------
-- Table structure for event_favorite
-- ----------------------------
DROP TABLE IF EXISTS "public"."event_favorite";
CREATE TABLE "public"."event_favorite" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "event_id" int8 NOT NULL,
  "notify_enabled" int2 NOT NULL DEFAULT 1,
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."event_favorite" IS 'User event favorite table';

-- ----------------------------
-- Records of event_favorite
-- ----------------------------
INSERT INTO "public"."event_favorite" VALUES (160000000000000001, 100000000000000001, 140000000000000001, 1, '2026-03-05 10:00:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000002, 100000000000000001, 140000000000000003, 1, '2026-03-08 15:30:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000003, 100000000000000002, 140000000000000002, 1, '2026-03-09 12:00:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000004, 100000000000000002, 140000000000000006, 0, '2026-03-16 09:00:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000005, 100000000000000003, 140000000000000007, 1, '2026-03-11 18:20:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000006, 100000000000000003, 140000000000000010, 1, '2026-03-17 13:10:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000007, 100000000000000004, 140000000000000005, 1, '2026-03-14 11:20:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000008, 100000000000000004, 140000000000000008, 1, '2026-03-19 09:30:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000009, 100000000000000005, 140000000000000003, 0, '2026-03-13 20:00:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000010, 100000000000000005, 140000000000000009, 1, '2026-03-20 10:10:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000011, 100000000000000006, 140000000000000011, 1, '2026-03-18 21:10:00');
INSERT INTO "public"."event_favorite" VALUES (160000000000000012, 100000000000000006, 140000000000000001, 0, '2026-03-20 21:10:00');

-- ----------------------------
-- Table structure for notification
-- ----------------------------
DROP TABLE IF EXISTS "public"."notification";
CREATE TABLE "public"."notification" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "type" int2 NOT NULL,
  "title" varchar(200) COLLATE "pg_catalog"."default" NOT NULL,
  "content" text COLLATE "pg_catalog"."default" NOT NULL,
  "is_read" int2 NOT NULL DEFAULT 0,
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."notification" IS 'In-app notification table';

-- ----------------------------
-- Records of notification
-- ----------------------------
INSERT INTO "public"."notification" VALUES (200000000000000001, 100000000000000001, 2, '支付成功', '您的订单 XM202603210001 已支付成功，请在演出开始前 90 分钟完成实名核验。', 0, '2026-03-21 19:59:00');
INSERT INTO "public"."notification" VALUES (200000000000000002, 100000000000000002, 1, '待支付提醒', '订单 XM202603210002 将于 2026-03-21 21:20 失效，请尽快完成支付。', 0, '2026-03-21 21:06:00');
INSERT INTO "public"."notification" VALUES (200000000000000003, 100000000000000003, 4, '订单已取消', '订单 XM202603200001 已取消，未支付座位库存已释放。', 1, '2026-03-20 11:11:00');
INSERT INTO "public"."notification" VALUES (200000000000000004, 100000000000000001, 2, '退款完成', '订单 XM202603100001 已原路退款 3100.00 元，请注意查收。', 0, '2026-03-20 16:21:00');
INSERT INTO "public"."notification" VALUES (200000000000000005, 100000000000000006, 3, '开票提醒已生效', '您关注的“2026 CBA 总决赛 G5 - 广州站”正在热售中。', 0, '2026-03-20 20:00:00');
INSERT INTO "public"."notification" VALUES (200000000000000006, 100000000000000005, 2, '电子票已核销', '广州城市马拉松嘉年华电子票已完成核销，感谢参与。', 1, '2025-11-09 12:50:00');

-- ----------------------------
-- Table structure for order_info
-- ----------------------------
DROP TABLE IF EXISTS "public"."order_info";
CREATE TABLE "public"."order_info" (
  "id" int8 NOT NULL,
  "order_no" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "user_id" int8 NOT NULL,
  "event_id" int8 NOT NULL,
  "ticket_tier_id" int8 NOT NULL,
  "quantity" int4 NOT NULL,
  "unit_price" numeric(10,2) NOT NULL,
  "total_amount" numeric(12,2) NOT NULL,
  "status" int2 NOT NULL DEFAULT 0,
  "cancel_reason" int2 NOT NULL DEFAULT 0,
  "pay_deadline" timestamp(3),
  "paid_at" timestamp(3),
  "cancelled_at" timestamp(3),
  "address_id" int8,
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."order_info" IS 'Order master table';

-- ----------------------------
-- Records of order_info
-- ----------------------------
INSERT INTO "public"."order_info" VALUES (2035904728679849984, '2035904724363911168', 2035026620841992192, 140000000000000001, 150000000000000001, 1, 1880.00, 1880.00, 2, 2, '2026-03-23 10:36:27.255', NULL, '2026-03-23 22:54:22.553', NULL, '2026-03-23 10:21:27.255', '2026-03-23 22:54:22.496');
INSERT INTO "public"."order_info" VALUES (2035904922406363136, '2035904917926846464', 2035026620841992192, 140000000000000002, 150000000000000004, 2, 1680.00, 3360.00, 2, 2, '2026-03-23 10:37:13.444', NULL, '2026-03-23 22:54:22.62', NULL, '2026-03-23 10:22:13.444', '2026-03-23 22:54:22.496');
INSERT INTO "public"."order_info" VALUES (2036094603374895104, '2036094598849241088', 2035026620841992192, 140000000000000009, 150000000000000024, 1, 1680.00, 1680.00, 2, 1, '2026-03-23 23:10:56.91', NULL, '2026-03-23 22:56:49.414', NULL, '2026-03-23 22:55:56.91', '2026-03-23 22:56:49.413');
INSERT INTO "public"."order_info" VALUES (2036094892618293248, '2036094888293965824', 2035026620841992192, 140000000000000001, 150000000000000001, 1, 1880.00, 1880.00, 3, 0, '2026-03-23 23:12:05.872', '2026-03-23 22:57:12', NULL, NULL, '2026-03-23 22:57:05.872', '2026-03-23 22:58:23.179');
INSERT INTO "public"."order_info" VALUES (2036095926619086848, '2036095922319925248', 2035026620841992192, 140000000000000002, 150000000000000004, 1, 1680.00, 1680.00, 3, 0, '2026-03-23 23:16:12.396', '2026-03-23 23:02:00', NULL, NULL, '2026-03-23 23:01:12.396', '2026-03-23 23:03:37.27');
INSERT INTO "public"."order_info" VALUES (2036096957532872704, '2036096953015607296', 2035026620841992192, 140000000000000001, 150000000000000001, 1, 1880.00, 1880.00, 3, 0, '2026-03-23 23:20:18.185', '2026-03-23 23:05:22', NULL, NULL, '2026-03-23 23:05:18.185', '2026-03-23 23:05:46.503');
INSERT INTO "public"."order_info" VALUES (2036098220622356480, '2036098216327389184', 2035026620841992192, 140000000000000002, 150000000000000004, 1, 1680.00, 1680.00, 2, 2, '2026-03-23 23:25:19.329', NULL, '2026-03-23 23:25:21.245', NULL, '2026-03-23 23:10:19.329', '2026-03-23 23:25:21.241');
INSERT INTO "public"."order_info" VALUES (2036102802261942272, '2036102797539155968', 2035026620841992192, 140000000000000008, 150000000000000021, 1, 1980.00, 1980.00, 3, 0, '2026-03-23 23:43:31.677', '2026-03-23 23:28:52', NULL, NULL, '2026-03-23 23:28:31.677', '2026-03-23 23:29:31.182');
INSERT INTO "public"."order_info" VALUES (2036107191462666240, '2036107186823766016', 2035026620841992192, 140000000000000008, 150000000000000023, 1, 780.00, 780.00, 3, 0, '2026-03-24 00:00:58.145', '2026-03-23 23:46:17', NULL, NULL, '2026-03-23 23:45:58.145', '2026-03-23 23:46:46.361');
INSERT INTO "public"."order_info" VALUES (2036114045567770624, '2036114040849178624', 2035026620841992192, 140000000000000009, 150000000000000024, 1, 1680.00, 1680.00, 4, 3, '2026-03-24 00:28:12.29', '2026-03-24 00:13:19', '2026-03-24 00:22:51.303', NULL, '2026-03-24 00:13:12.29', '2026-03-24 00:22:51.299');
INSERT INTO "public"."order_info" VALUES (2036116653820878848, '2036116649517522944', 2035026620841992192, 140000000000000002, 150000000000000004, 1, 1680.00, 1680.00, 4, 3, '2026-03-24 00:38:34.147', '2026-03-24 00:23:41', '2026-03-24 00:24:27.275', NULL, '2026-03-24 00:23:34.147', '2026-03-24 00:24:27.27');
INSERT INTO "public"."order_info" VALUES (2036126401207214080, '2036126396660588544', 2035026620841992192, 140000000000000010, 150000000000000029, 2, 280.00, 560.00, 4, 3, '2026-03-24 01:17:18.104', '2026-03-24 01:16:32', '2026-03-24 01:17:23.567', 2035038354281603072, '2026-03-24 01:02:18.104', '2026-03-24 01:17:23.56');

-- ----------------------------
-- Table structure for order_ticket
-- ----------------------------
DROP TABLE IF EXISTS "public"."order_ticket";
CREATE TABLE "public"."order_ticket" (
  "id" int8 NOT NULL,
  "order_id" int8 NOT NULL,
  "ticket_buyer_id" int8 NOT NULL,
  "ticket_code" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "qr_code_url" varchar(255) COLLATE "pg_catalog"."default",
  "status" int2 NOT NULL DEFAULT 0,
  "seat_info" varchar(255) COLLATE "pg_catalog"."default",
  "verified_at" timestamp(3),
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."order_ticket" IS 'Electronic ticket detail table';

-- ----------------------------
-- Records of order_ticket
-- ----------------------------
INSERT INTO "public"."order_ticket" VALUES (2035904728696627200, 2035904728679849984, 2035038076035670016, 'TK203590472436391116801', 'mock://ticket/TK203590472436391116801', 2, '', NULL, '2026-03-23 10:21:27.255', '2026-03-23 22:54:22.496');
INSERT INTO "public"."order_ticket" VALUES (2035904922498637824, 2035904922406363136, 2035038076035670016, 'TK203590491792684646401', 'mock://ticket/TK203590491792684646401', 2, '', NULL, '2026-03-23 10:22:13.444', '2026-03-23 22:54:22.496');
INSERT INTO "public"."order_ticket" VALUES (2035904922498637825, 2035904922406363136, 2035038227605233664, 'TK203590491792684646402', 'mock://ticket/TK203590491792684646402', 2, '', NULL, '2026-03-23 10:22:13.444', '2026-03-23 22:54:22.496');
INSERT INTO "public"."order_ticket" VALUES (2036094603387478016, 2036094603374895104, 2035038076035670016, 'TK203609459884924108801', 'mock://ticket/TK203609459884924108801', 2, '', NULL, '2026-03-23 22:55:56.91', '2026-03-23 22:56:49.413');
INSERT INTO "public"."order_ticket" VALUES (2036094892643459072, 2036094892618293248, 2035038076035670016, 'TK203609488829396582401', 'mock://ticket/TK203609488829396582401', 0, '', NULL, '2026-03-23 22:57:05.872', '2026-03-23 22:57:05.872');
INSERT INTO "public"."order_ticket" VALUES (2036095926623281152, 2036095926619086848, 2035038076035670016, 'TK203609592231992524801', 'mock://ticket/TK203609592231992524801', 0, '', NULL, '2026-03-23 23:01:12.396', '2026-03-23 23:01:12.396');
INSERT INTO "public"."order_ticket" VALUES (2036096957553844224, 2036096957532872704, 2035038076035670016, 'TK203609695301560729601', 'mock://ticket/TK203609695301560729601', 0, '', NULL, '2026-03-23 23:05:18.185', '2026-03-23 23:05:18.185');
INSERT INTO "public"."order_ticket" VALUES (2036098220630745088, 2036098220622356480, 2035038076035670016, 'TK203609821632738918401', 'mock://ticket/TK203609821632738918401', 2, '', NULL, '2026-03-23 23:10:19.329', '2026-03-23 23:25:21.241');
INSERT INTO "public"."order_ticket" VALUES (2036102802291302400, 2036102802261942272, 2035038076035670016, 'TK203610279753915596801', 'mock://ticket/TK203610279753915596801', 0, '', NULL, '2026-03-23 23:28:31.677', '2026-03-23 23:28:31.677');
INSERT INTO "public"."order_ticket" VALUES (2036107191601078272, 2036107191462666240, 2035038076035670016, 'TK203610718682376601601', 'mock://ticket/TK203610718682376601601', 0, '', NULL, '2026-03-23 23:45:58.145', '2026-03-23 23:45:58.145');
INSERT INTO "public"."order_ticket" VALUES (2036114045588742144, 2036114045567770624, 2035038076035670016, 'TK203611404084917862401', 'mock://ticket/TK203611404084917862401', 2, '', NULL, '2026-03-24 00:13:12.29', '2026-03-24 00:22:51.299');
INSERT INTO "public"."order_ticket" VALUES (2036116653833461760, 2036116653820878848, 2035038076035670016, 'TK203611664951752294401', 'mock://ticket/TK203611664951752294401', 2, '', NULL, '2026-03-24 00:23:34.147', '2026-03-24 00:24:27.27');
INSERT INTO "public"."order_ticket" VALUES (2036126401232379904, 2036126401207214080, 2035038076035670016, 'TK203612639666058854401', 'mock://ticket/TK203612639666058854401', 2, '', NULL, '2026-03-24 01:02:18.104', '2026-03-24 01:17:23.56');
INSERT INTO "public"."order_ticket" VALUES (2036126401232379905, 2036126401207214080, 2035038227605233664, 'TK203612639666058854402', 'mock://ticket/TK203612639666058854402', 2, '', NULL, '2026-03-24 01:02:18.104', '2026-03-24 01:17:23.56');

-- ----------------------------
-- Table structure for payment
-- ----------------------------
DROP TABLE IF EXISTS "public"."payment";
CREATE TABLE "public"."payment" (
  "id" int8 NOT NULL,
  "payment_no" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "order_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "pay_method" int2 NOT NULL,
  "amount" numeric(12,2) NOT NULL,
  "status" int2 NOT NULL DEFAULT 0,
  "trade_no" varchar(128) COLLATE "pg_catalog"."default",
  "paid_at" timestamp(3),
  "callback_data" text COLLATE "pg_catalog"."default",
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."payment" IS 'Payment attempt table';

-- ----------------------------
-- Records of payment
-- ----------------------------
INSERT INTO "public"."payment" VALUES (2035904728713404416, 'P2035904728713404417', 2035904728679849984, 2035026620841992192, 1, 1880.00, 2, NULL, NULL, NULL, '2026-03-23 10:21:27.255', '2026-03-23 22:54:22.496');
INSERT INTO "public"."payment" VALUES (2035904922540580864, 'P2035904922540580865', 2035904922406363136, 2035026620841992192, 1, 3360.00, 2, NULL, NULL, NULL, '2026-03-23 10:22:13.444', '2026-03-23 22:54:22.496');
INSERT INTO "public"."payment" VALUES (2036094603404255232, 'P2036094603404255233', 2036094603374895104, 2035026620841992192, 1, 1680.00, 2, NULL, NULL, NULL, '2026-03-23 22:55:56.91', '2026-03-23 22:56:49.413');
INSERT INTO "public"."payment" VALUES (2036094892656041984, 'P2036094892656041985', 2036094892618293248, 2035026620841992192, 1, 1880.00, 1, 'cs_test_a1Zp5OiIu3I0jb1KYOyYuLWS0tRam0MwvmLKXryAMzT0PBAa24QSY3G6d1', '2026-03-23 22:57:12', '{
  "id": "evt_1TE9vvPm68qiSwnQ7dIeur64",
  "object": "event",
  "api_version": "2025-11-17.clover",
  "created": 1774277903,
  "data": {
    "object": {
      "id": "cs_test_a1Zp5OiIu3I0jb1KYOyYuLWS0tRam0MwvmLKXryAMzT0PBAa24QSY3G6d1",
      "object": "checkout.session",
      "adaptive_pricing": {
        "enabled": true
      },
      "after_expiration": null,
      "allow_promotion_codes": null,
      "amount_subtotal": 188000,
      "amount_total": 188000,
      "automatic_tax": {
        "enabled": false,
        "liability": null,
        "provider": null,
        "status": null
      },
      "billing_address_collection": null,
      "branding_settings": {
        "background_color": "#ffffff",
        "border_style": "rounded",
        "button_color": "#0074d4",
        "display_name": "New business 沙盒",
        "font_family": "default",
        "icon": null,
        "logo": null
      },
      "cancel_url": "http://127.0.0.1:3000/order/list?orderNo=2036094888293965824\u0026paymentNo=P2036094892656041985\u0026paymentResult=cancel",
      "client_reference_id": "P2036094892656041985",
      "client_secret": null,
      "collected_information": null,
      "consent": null,
      "consent_collection": null,
      "created": 1774277832,
      "currency": "cny",
      "currency_conversion": null,
      "custom_fields": [],
      "custom_text": {
        "after_submit": null,
        "shipping_address": null,
        "submit": null,
        "terms_of_service_acceptance": null
      },
      "customer": null,
      "customer_account": null,
      "customer_creation": "if_required",
      "customer_details": {
        "address": {
          "city": null,
          "country": "HK",
          "line1": null,
          "line2": null,
          "postal_code": null,
          "state": null
        },
        "business_name": null,
        "email": "lihua1823494260@gmail.com",
        "individual_name": null,
        "name": "123",
        "phone": null,
        "tax_exempt": "none",
        "tax_ids": []
      },
      "customer_email": null,
      "discounts": [],
      "expires_at": 1774279631,
      "integration_identifier": null,
      "invoice": null,
      "invoice_creation": {
        "enabled": false,
        "invoice_data": {
          "account_tax_ids": null,
          "custom_fields": null,
          "description": null,
          "footer": null,
          "issuer": null,
          "metadata": {},
          "rendering_options": null
        }
      },
      "livemode": false,
      "locale": null,
      "metadata": {
        "user_id": "2035026620841992192",
        "payment_no": "P2036094892656041985",
        "order_no": "2036094888293965824"
      },
      "mode": "payment",
      "origin_context": null,
      "payment_intent": "pi_3TE9vuPm68qiSwnQ0g3sMmrO",
      "payment_link": null,
      "payment_method_collection": "if_required",
      "payment_method_configuration_details": null,
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "payment_status": "paid",
      "permissions": null,
      "phone_number_collection": {
        "enabled": false
      },
      "recovered_from": null,
      "saved_payment_method_options": null,
      "setup_intent": null,
      "shipping_address_collection": null,
      "shipping_cost": null,
      "shipping_options": [],
      "status": "complete",
      "submit_type": null,
      "subscription": null,
      "success_url": "http://127.0.0.1:3000/order/list?orderNo=2036094888293965824\u0026paymentNo=P2036094892656041985\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D",
      "total_details": {
        "amount_discount": 0,
        "amount_shipping": 0,
        "amount_tax": 0
      },
      "ui_mode": "hosted",
      "url": null,
      "wallet_options": null
    }
  },
  "livemode": false,
  "pending_webhooks": 2,
  "request": {
    "id": null,
    "idempotency_key": null
  },
  "type": "checkout.session.completed"
}', '2026-03-23 22:57:05.872', '2026-03-23 22:58:23.179');
INSERT INTO "public"."payment" VALUES (2036095926635864064, 'P2036095926635864065', 2036095926619086848, 2035026620841992192, 1, 1680.00, 1, 'cs_test_a1V4DuJ2CnQMnoZPpHz5kfBZgIuMNdxvgf5nW9x9fb7s0DnYU1oXivEeul', '2026-03-23 23:02:00', '{
  "id": "evt_1TEA0zPm68qiSwnQaXZLaHB6",
  "object": "event",
  "api_version": "2025-11-17.clover",
  "created": 1774278217,
  "data": {
    "object": {
      "id": "cs_test_a1V4DuJ2CnQMnoZPpHz5kfBZgIuMNdxvgf5nW9x9fb7s0DnYU1oXivEeul",
      "object": "checkout.session",
      "adaptive_pricing": {
        "enabled": true
      },
      "after_expiration": null,
      "allow_promotion_codes": null,
      "amount_subtotal": 168000,
      "amount_total": 168000,
      "automatic_tax": {
        "enabled": false,
        "liability": null,
        "provider": null,
        "status": null
      },
      "billing_address_collection": null,
      "branding_settings": {
        "background_color": "#ffffff",
        "border_style": "rounded",
        "button_color": "#0074d4",
        "display_name": "New business 沙盒",
        "font_family": "default",
        "icon": null,
        "logo": null
      },
      "cancel_url": "http://127.0.0.1:3000/order/list?orderNo=2036095922319925248\u0026paymentNo=P2036095926635864065\u0026paymentResult=cancel",
      "client_reference_id": "P2036095926635864065",
      "client_secret": null,
      "collected_information": null,
      "consent": null,
      "consent_collection": null,
      "created": 1774278120,
      "currency": "cny",
      "currency_conversion": null,
      "custom_fields": [],
      "custom_text": {
        "after_submit": null,
        "shipping_address": null,
        "submit": null,
        "terms_of_service_acceptance": null
      },
      "customer": null,
      "customer_account": null,
      "customer_creation": "if_required",
      "customer_details": {
        "address": {
          "city": null,
          "country": "HK",
          "line1": null,
          "line2": null,
          "postal_code": null,
          "state": null
        },
        "business_name": null,
        "email": "lihua1823494260@gmail.com",
        "individual_name": null,
        "name": "123",
        "phone": null,
        "tax_exempt": "none",
        "tax_ids": []
      },
      "customer_email": null,
      "discounts": [],
      "expires_at": 1774279919,
      "integration_identifier": null,
      "invoice": null,
      "invoice_creation": {
        "enabled": false,
        "invoice_data": {
          "account_tax_ids": null,
          "custom_fields": null,
          "description": null,
          "footer": null,
          "issuer": null,
          "metadata": {},
          "rendering_options": null
        }
      },
      "livemode": false,
      "locale": null,
      "metadata": {
        "user_id": "2035026620841992192",
        "payment_no": "P2036095926635864065",
        "order_no": "2036095922319925248"
      },
      "mode": "payment",
      "origin_context": null,
      "payment_intent": "pi_3TEA0yPm68qiSwnQ15mgWFYG",
      "payment_link": null,
      "payment_method_collection": "if_required",
      "payment_method_configuration_details": null,
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "payment_status": "paid",
      "permissions": null,
      "phone_number_collection": {
        "enabled": false
      },
      "recovered_from": null,
      "saved_payment_method_options": null,
      "setup_intent": null,
      "shipping_address_collection": null,
      "shipping_cost": null,
      "shipping_options": [],
      "status": "complete",
      "submit_type": null,
      "subscription": null,
      "success_url": "http://127.0.0.1:3000/order/list?orderNo=2036095922319925248\u0026paymentNo=P2036095926635864065\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D",
      "total_details": {
        "amount_discount": 0,
        "amount_shipping": 0,
        "amount_tax": 0
      },
      "ui_mode": "hosted",
      "url": null,
      "wallet_options": null
    }
  },
  "livemode": false,
  "pending_webhooks": 2,
  "request": {
    "id": null,
    "idempotency_key": null
  },
  "type": "checkout.session.completed"
}', '2026-03-23 23:01:12.396', '2026-03-23 23:03:37.27');
INSERT INTO "public"."payment" VALUES (2036096957583204352, 'P2036096957583204353', 2036096957532872704, 2035026620841992192, 1, 1880.00, 1, 'cs_test_a1AYig1ktZXxE0jzw2c2RNeoJ5KeTajVX2Ng8dGvs6pFFYZWmKK5pvVYrI', '2026-03-23 23:05:22', '{
  "id": "evt_1TEA35Pm68qiSwnQN07UZSO7",
  "object": "event",
  "api_version": "2025-11-17.clover",
  "created": 1774278346,
  "data": {
    "object": {
      "id": "cs_test_a1AYig1ktZXxE0jzw2c2RNeoJ5KeTajVX2Ng8dGvs6pFFYZWmKK5pvVYrI",
      "object": "checkout.session",
      "adaptive_pricing": {
        "enabled": true
      },
      "after_expiration": null,
      "allow_promotion_codes": null,
      "amount_subtotal": 188000,
      "amount_total": 188000,
      "automatic_tax": {
        "enabled": false,
        "liability": null,
        "provider": null,
        "status": null
      },
      "billing_address_collection": null,
      "branding_settings": {
        "background_color": "#ffffff",
        "border_style": "rounded",
        "button_color": "#0074d4",
        "display_name": "New business 沙盒",
        "font_family": "default",
        "icon": null,
        "logo": null
      },
      "cancel_url": "http://localhost:9998/order/list?orderNo=2036096953015607296\u0026paymentNo=P2036096957583204353\u0026paymentResult=cancel",
      "client_reference_id": "P2036096957583204353",
      "client_secret": null,
      "collected_information": null,
      "consent": null,
      "consent_collection": null,
      "created": 1774278322,
      "currency": "cny",
      "currency_conversion": null,
      "custom_fields": [],
      "custom_text": {
        "after_submit": null,
        "shipping_address": null,
        "submit": null,
        "terms_of_service_acceptance": null
      },
      "customer": null,
      "customer_account": null,
      "customer_creation": "if_required",
      "customer_details": {
        "address": {
          "city": null,
          "country": "HK",
          "line1": null,
          "line2": null,
          "postal_code": null,
          "state": null
        },
        "business_name": null,
        "email": "lihua1823494260@gmail.com",
        "individual_name": null,
        "name": "123",
        "phone": null,
        "tax_exempt": "none",
        "tax_ids": []
      },
      "customer_email": null,
      "discounts": [],
      "expires_at": 1774280120,
      "integration_identifier": null,
      "invoice": null,
      "invoice_creation": {
        "enabled": false,
        "invoice_data": {
          "account_tax_ids": null,
          "custom_fields": null,
          "description": null,
          "footer": null,
          "issuer": null,
          "metadata": {},
          "rendering_options": null
        }
      },
      "livemode": false,
      "locale": null,
      "metadata": {
        "user_id": "2035026620841992192",
        "payment_no": "P2036096957583204353",
        "order_no": "2036096953015607296"
      },
      "mode": "payment",
      "origin_context": null,
      "payment_intent": "pi_3TEA32Pm68qiSwnQ0s7ueAXD",
      "payment_link": null,
      "payment_method_collection": "if_required",
      "payment_method_configuration_details": null,
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "payment_status": "paid",
      "permissions": null,
      "phone_number_collection": {
        "enabled": false
      },
      "recovered_from": null,
      "saved_payment_method_options": null,
      "setup_intent": null,
      "shipping_address_collection": null,
      "shipping_cost": null,
      "shipping_options": [],
      "status": "complete",
      "submit_type": null,
      "subscription": null,
      "success_url": "http://localhost:9998/order/list?orderNo=2036096953015607296\u0026paymentNo=P2036096957583204353\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D",
      "total_details": {
        "amount_discount": 0,
        "amount_shipping": 0,
        "amount_tax": 0
      },
      "ui_mode": "hosted",
      "url": null,
      "wallet_options": null
    }
  },
  "livemode": false,
  "pending_webhooks": 2,
  "request": {
    "id": null,
    "idempotency_key": null
  },
  "type": "checkout.session.completed"
}', '2026-03-23 23:05:18.185', '2026-03-23 23:05:46.503');
INSERT INTO "public"."payment" VALUES (2036098220639133696, 'P2036098220639133697', 2036098220622356480, 2035026620841992192, 1, 1680.00, 2, 'cs_test_a17lyZuw9bkMsgK94mSNJUEz0sZerCHzFG4eyQkOUeuj1Y50tLgz8EC85h', NULL, '{"adaptive_pricing":{"enabled":true},"after_expiration":null,"allow_promotion_codes":false,"amount_subtotal":168000,"amount_total":168000,"automatic_tax":{"enabled":false,"liability":null,"provider":"","status":""},"billing_address_collection":"","branding_settings":{"background_color":"#ffffff","border_style":"rounded","button_color":"#0074d4","display_name":"New business 沙盒","font_family":"default","icon":null,"logo":null},"cancel_url":"http://localhost:9998/order/list?orderNo=2036098216327389184\u0026paymentNo=P2036098220639133697\u0026paymentResult=cancel","client_reference_id":"P2036098220639133697","client_secret":"","collected_information":null,"consent":null,"consent_collection":null,"created":1774278638,"currency":"cny","currency_conversion":null,"customer":null,"customer_account":"","customer_creation":"if_required","customer_details":null,"customer_email":"","custom_fields":[],"custom_text":{"after_submit":null,"shipping_address":null,"submit":null,"terms_of_service_acceptance":null},"discounts":[],"excluded_payment_method_types":null,"expires_at":1774280437,"id":"cs_test_a17lyZuw9bkMsgK94mSNJUEz0sZerCHzFG4eyQkOUeuj1Y50tLgz8EC85h","invoice":null,"invoice_creation":{"enabled":false,"invoice_data":{"account_tax_ids":null,"custom_fields":null,"description":"","footer":"","issuer":null,"metadata":{},"rendering_options":null}},"line_items":null,"livemode":false,"locale":"","metadata":{"order_no":"2036098216327389184","payment_no":"P2036098220639133697","user_id":"2035026620841992192"},"mode":"payment","name_collection":null,"object":"checkout.session","optional_items":null,"origin_context":"","payment_intent":null,"payment_link":null,"payment_method_collection":"if_required","payment_method_configuration_details":null,"payment_method_options":{"acss_debit":null,"affirm":null,"afterpay_clearpay":null,"alipay":null,"alma":null,"amazon_pay":null,"au_becs_debit":null,"bacs_debit":null,"bancontact":null,"billie":null,"boleto":null,"card":{"capture_method":"","installments":null,"request_extended_authorization":"","request_incremental_authorization":"","request_multicapture":"","request_overcapture":"","request_three_d_secure":"automatic","restrictions":null,"setup_future_usage":"","statement_descriptor_suffix_kana":"","statement_descriptor_suffix_kanji":""},"cashapp":null,"customer_balance":null,"eps":null,"fpx":null,"giropay":null,"grabpay":null,"ideal":null,"kakao_pay":null,"klarna":null,"konbini":null,"kr_card":null,"link":null,"mobilepay":null,"multibanco":null,"naver_pay":null,"oxxo":null,"p24":null,"payco":null,"paynow":null,"paypal":null,"payto":null,"pix":null,"revolut_pay":null,"samsung_pay":null,"satispay":null,"sepa_debit":null,"sofort":null,"swish":null,"twint":null,"us_bank_account":null},"payment_method_types":["card"],"payment_status":"unpaid","permissions":null,"phone_number_collection":{"enabled":false},"presentment_details":null,"recovered_from":"","redirect_on_completion":"","return_url":"","saved_payment_method_options":null,"setup_intent":null,"shipping_address_collection":null,"shipping_cost":null,"shipping_options":[],"status":"open","submit_type":"","subscription":null,"success_url":"http://localhost:9998/order/list?orderNo=2036098216327389184\u0026paymentNo=P2036098220639133697\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D","tax_id_collection":null,"total_details":{"amount_discount":0,"amount_shipping":0,"amount_tax":0,"breakdown":null},"ui_mode":"hosted","url":"https://checkout.stripe.com/c/pay/cs_test_a17lyZuw9bkMsgK94mSNJUEz0sZerCHzFG4eyQkOUeuj1Y50tLgz8EC85h#fidnandhYHdWcXxpYCc%2FJ2FgY2RwaXEnKSdkdWxOYHwnPyd1blpxYHZxWjA0VlJiRGhVaDM9dGxWcmtUPW5EbTVPXzZrfEtvRmZ3XExnQ0AxRHRgVkM9Z2F3REtnUzVcV0tXUT1gU2F1aHY1cm9fUTFCTEJxS1BgcXViS1dKTE80VUZPNTVOcm1mVlwyVCcpJ2N3amhWYHdzYHcnP3F3cGApJ2dkZm5id2pwa2FGamlqdyc%2FJyZjY2NjY2MnKSdpZHxqcHFRfHVgJz8ndmxrYmlgWmxxYGgnKSdga2RnaWBVaWRmYG1qaWFgd3YnP3F3cGB4JSUl","wallet_options":null}', '2026-03-23 23:10:19.329', '2026-03-23 23:25:21.241');
INSERT INTO "public"."payment" VALUES (2036102802320662528, 'P2036102802320662529', 2036102802261942272, 2035026620841992192, 1, 1980.00, 1, 'cs_test_a1xZznWSKrZXxxNGYnS9XtYrGAwIpNj5bTzq02ovB0qAfrjvrsAtv9H9Az', '2026-03-23 23:28:52', '{
  "id": "evt_1TEAQ3Pm68qiSwnQJPen1ink",
  "object": "event",
  "api_version": "2025-11-17.clover",
  "created": 1774279771,
  "data": {
    "object": {
      "id": "cs_test_a1xZznWSKrZXxxNGYnS9XtYrGAwIpNj5bTzq02ovB0qAfrjvrsAtv9H9Az",
      "object": "checkout.session",
      "adaptive_pricing": {
        "enabled": true
      },
      "after_expiration": null,
      "allow_promotion_codes": null,
      "amount_subtotal": 198000,
      "amount_total": 198000,
      "automatic_tax": {
        "enabled": false,
        "liability": null,
        "provider": null,
        "status": null
      },
      "billing_address_collection": null,
      "branding_settings": {
        "background_color": "#ffffff",
        "border_style": "rounded",
        "button_color": "#0074d4",
        "display_name": "New business 沙盒",
        "font_family": "default",
        "icon": null,
        "logo": null
      },
      "cancel_url": "http://localhost:9998/order/list?orderNo=2036102797539155968\u0026paymentNo=P2036102802320662529\u0026paymentResult=cancel",
      "client_reference_id": "P2036102802320662529",
      "client_secret": null,
      "collected_information": null,
      "consent": null,
      "consent_collection": null,
      "created": 1774279732,
      "currency": "cny",
      "currency_conversion": null,
      "custom_fields": [],
      "custom_text": {
        "after_submit": null,
        "shipping_address": null,
        "submit": null,
        "terms_of_service_acceptance": null
      },
      "customer": null,
      "customer_account": null,
      "customer_creation": "if_required",
      "customer_details": {
        "address": {
          "city": null,
          "country": "HK",
          "line1": null,
          "line2": null,
          "postal_code": null,
          "state": null
        },
        "business_name": null,
        "email": "lihua1823494260@gmail.com",
        "individual_name": null,
        "name": "123",
        "phone": null,
        "tax_exempt": "none",
        "tax_ids": []
      },
      "customer_email": null,
      "discounts": [],
      "expires_at": 1774281530,
      "integration_identifier": null,
      "invoice": null,
      "invoice_creation": {
        "enabled": false,
        "invoice_data": {
          "account_tax_ids": null,
          "custom_fields": null,
          "description": null,
          "footer": null,
          "issuer": null,
          "metadata": {},
          "rendering_options": null
        }
      },
      "livemode": false,
      "locale": null,
      "metadata": {
        "user_id": "2035026620841992192",
        "payment_no": "P2036102802320662529",
        "order_no": "2036102797539155968"
      },
      "mode": "payment",
      "origin_context": null,
      "payment_intent": "pi_3TEAQ1Pm68qiSwnQ0DoYM5lX",
      "payment_link": null,
      "payment_method_collection": "if_required",
      "payment_method_configuration_details": null,
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "payment_status": "paid",
      "permissions": null,
      "phone_number_collection": {
        "enabled": false
      },
      "recovered_from": null,
      "saved_payment_method_options": null,
      "setup_intent": null,
      "shipping_address_collection": null,
      "shipping_cost": null,
      "shipping_options": [],
      "status": "complete",
      "submit_type": null,
      "subscription": null,
      "success_url": "http://localhost:9998/payment/processing?orderNo=2036102797539155968\u0026paymentNo=P2036102802320662529\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D",
      "total_details": {
        "amount_discount": 0,
        "amount_shipping": 0,
        "amount_tax": 0
      },
      "ui_mode": "hosted",
      "url": null,
      "wallet_options": null
    }
  },
  "livemode": false,
  "pending_webhooks": 2,
  "request": {
    "id": null,
    "idempotency_key": null
  },
  "type": "checkout.session.completed"
}', '2026-03-23 23:28:31.677', '2026-03-23 23:29:31.182');
INSERT INTO "public"."payment" VALUES (2036107191718518784, 'P2036107191718518785', 2036107191462666240, 2035026620841992192, 1, 780.00, 1, 'cs_test_a1VQoQrwCfpDpufYuTEmAsii1Kx0aoInIwiYtswURa9G7zjLnPZbsf5N1N', '2026-03-23 23:46:17', '{
  "id": "evt_1TEAgkPm68qiSwnQLb7f5Frt",
  "object": "event",
  "api_version": "2025-11-17.clover",
  "created": 1774280806,
  "data": {
    "object": {
      "id": "cs_test_a1VQoQrwCfpDpufYuTEmAsii1Kx0aoInIwiYtswURa9G7zjLnPZbsf5N1N",
      "object": "checkout.session",
      "adaptive_pricing": {
        "enabled": true
      },
      "after_expiration": null,
      "allow_promotion_codes": null,
      "amount_subtotal": 78000,
      "amount_total": 78000,
      "automatic_tax": {
        "enabled": false,
        "liability": null,
        "provider": null,
        "status": null
      },
      "billing_address_collection": null,
      "branding_settings": {
        "background_color": "#ffffff",
        "border_style": "rounded",
        "button_color": "#0074d4",
        "display_name": "New business 沙盒",
        "font_family": "default",
        "icon": null,
        "logo": null
      },
      "cancel_url": "http://localhost:9998/order/list?orderNo=2036107186823766016\u0026paymentNo=P2036107191718518785\u0026paymentResult=cancel",
      "client_reference_id": "P2036107191718518785",
      "client_secret": null,
      "collected_information": null,
      "consent": null,
      "consent_collection": null,
      "created": 1774280777,
      "currency": "cny",
      "currency_conversion": null,
      "custom_fields": [],
      "custom_text": {
        "after_submit": null,
        "shipping_address": null,
        "submit": null,
        "terms_of_service_acceptance": null
      },
      "customer": null,
      "customer_account": null,
      "customer_creation": "if_required",
      "customer_details": {
        "address": {
          "city": null,
          "country": "HK",
          "line1": null,
          "line2": null,
          "postal_code": null,
          "state": null
        },
        "business_name": null,
        "email": "lihua1823494260@gmail.com",
        "individual_name": null,
        "name": "123",
        "phone": null,
        "tax_exempt": "none",
        "tax_ids": []
      },
      "customer_email": null,
      "discounts": [],
      "expires_at": 1774282575,
      "integration_identifier": null,
      "invoice": null,
      "invoice_creation": {
        "enabled": false,
        "invoice_data": {
          "account_tax_ids": null,
          "custom_fields": null,
          "description": null,
          "footer": null,
          "issuer": null,
          "metadata": {},
          "rendering_options": null
        }
      },
      "livemode": false,
      "locale": null,
      "metadata": {
        "user_id": "2035026620841992192",
        "payment_no": "P2036107191718518785",
        "order_no": "2036107186823766016"
      },
      "mode": "payment",
      "origin_context": null,
      "payment_intent": "pi_3TEAgjPm68qiSwnQ1saGgWRC",
      "payment_link": null,
      "payment_method_collection": "if_required",
      "payment_method_configuration_details": null,
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "payment_status": "paid",
      "permissions": null,
      "phone_number_collection": {
        "enabled": false
      },
      "recovered_from": null,
      "saved_payment_method_options": null,
      "setup_intent": null,
      "shipping_address_collection": null,
      "shipping_cost": null,
      "shipping_options": [],
      "status": "complete",
      "submit_type": null,
      "subscription": null,
      "success_url": "http://localhost:9998/payment/processing?orderNo=2036107186823766016\u0026paymentNo=P2036107191718518785\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D",
      "total_details": {
        "amount_discount": 0,
        "amount_shipping": 0,
        "amount_tax": 0
      },
      "ui_mode": "hosted",
      "url": null,
      "wallet_options": null
    }
  },
  "livemode": false,
  "pending_webhooks": 2,
  "request": {
    "id": null,
    "idempotency_key": null
  },
  "type": "checkout.session.completed"
}', '2026-03-23 23:45:58.145', '2026-03-23 23:46:46.361');
INSERT INTO "public"."payment" VALUES (2036114045622296576, 'P2036114045622296577', 2036114045567770624, 2035026620841992192, 1, 1680.00, 1, 'cs_test_a1A0XlKQ7D4fPElnnMJSFJdQzUvmzcSHbiyOBzi9A8gMbGGNL0f9rU2vz0', '2026-03-24 00:13:19', '{
  "id": "evt_1TEB72Pm68qiSwnQcTSGhJVU",
  "object": "event",
  "api_version": "2025-11-17.clover",
  "created": 1774282436,
  "data": {
    "object": {
      "id": "cs_test_a1A0XlKQ7D4fPElnnMJSFJdQzUvmzcSHbiyOBzi9A8gMbGGNL0f9rU2vz0",
      "object": "checkout.session",
      "adaptive_pricing": {
        "enabled": true
      },
      "after_expiration": null,
      "allow_promotion_codes": null,
      "amount_subtotal": 168000,
      "amount_total": 168000,
      "automatic_tax": {
        "enabled": false,
        "liability": null,
        "provider": null,
        "status": null
      },
      "billing_address_collection": null,
      "branding_settings": {
        "background_color": "#ffffff",
        "border_style": "rounded",
        "button_color": "#0074d4",
        "display_name": "New business 沙盒",
        "font_family": "default",
        "icon": null,
        "logo": null
      },
      "cancel_url": "http://localhost:9998/order/list?orderNo=2036114040849178624\u0026paymentNo=P2036114045622296577\u0026paymentResult=cancel",
      "client_reference_id": "P2036114045622296577",
      "client_secret": null,
      "collected_information": null,
      "consent": null,
      "consent_collection": null,
      "created": 1774282399,
      "currency": "cny",
      "currency_conversion": null,
      "custom_fields": [],
      "custom_text": {
        "after_submit": null,
        "shipping_address": null,
        "submit": null,
        "terms_of_service_acceptance": null
      },
      "customer": null,
      "customer_account": null,
      "customer_creation": "if_required",
      "customer_details": {
        "address": {
          "city": null,
          "country": "HK",
          "line1": null,
          "line2": null,
          "postal_code": null,
          "state": null
        },
        "business_name": null,
        "email": "lihua1823494260@gmail.com",
        "individual_name": null,
        "name": "123",
        "phone": null,
        "tax_exempt": "none",
        "tax_ids": []
      },
      "customer_email": null,
      "discounts": [],
      "expires_at": 1774284198,
      "integration_identifier": null,
      "invoice": null,
      "invoice_creation": {
        "enabled": false,
        "invoice_data": {
          "account_tax_ids": null,
          "custom_fields": null,
          "description": null,
          "footer": null,
          "issuer": null,
          "metadata": {},
          "rendering_options": null
        }
      },
      "livemode": false,
      "locale": null,
      "metadata": {
        "user_id": "2035026620841992192",
        "payment_no": "P2036114045622296577",
        "order_no": "2036114040849178624"
      },
      "mode": "payment",
      "origin_context": null,
      "payment_intent": "pi_3TEB71Pm68qiSwnQ0foS7I7V",
      "payment_link": null,
      "payment_method_collection": "if_required",
      "payment_method_configuration_details": null,
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "payment_status": "paid",
      "permissions": null,
      "phone_number_collection": {
        "enabled": false
      },
      "recovered_from": null,
      "saved_payment_method_options": null,
      "setup_intent": null,
      "shipping_address_collection": null,
      "shipping_cost": null,
      "shipping_options": [],
      "status": "complete",
      "submit_type": null,
      "subscription": null,
      "success_url": "http://localhost:9998/payment/processing?orderNo=2036114040849178624\u0026paymentNo=P2036114045622296577\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D",
      "total_details": {
        "amount_discount": 0,
        "amount_shipping": 0,
        "amount_tax": 0
      },
      "ui_mode": "hosted",
      "url": null,
      "wallet_options": null
    }
  },
  "livemode": false,
  "pending_webhooks": 2,
  "request": {
    "id": null,
    "idempotency_key": null
  },
  "type": "checkout.session.completed"
}', '2026-03-24 00:13:12.29', '2026-03-24 00:13:56.054');
INSERT INTO "public"."payment" VALUES (2036116653846044672, 'P2036116653846044673', 2036116653820878848, 2035026620841992192, 1, 1680.00, 1, 'cs_test_a1UlJUZ0WtOuUX52HPtI5s2jIpnSwATnqqLd7VD9BCY0wLYLmm4WEin504', '2026-03-24 00:23:41', '{
  "id": "evt_1TEBGxPm68qiSwnQXLv6rMKf",
  "object": "event",
  "api_version": "2025-11-17.clover",
  "created": 1774283050,
  "data": {
    "object": {
      "id": "cs_test_a1UlJUZ0WtOuUX52HPtI5s2jIpnSwATnqqLd7VD9BCY0wLYLmm4WEin504",
      "object": "checkout.session",
      "adaptive_pricing": {
        "enabled": true
      },
      "after_expiration": null,
      "allow_promotion_codes": null,
      "amount_subtotal": 168000,
      "amount_total": 168000,
      "automatic_tax": {
        "enabled": false,
        "liability": null,
        "provider": null,
        "status": null
      },
      "billing_address_collection": null,
      "branding_settings": {
        "background_color": "#ffffff",
        "border_style": "rounded",
        "button_color": "#0074d4",
        "display_name": "New business 沙盒",
        "font_family": "default",
        "icon": null,
        "logo": null
      },
      "cancel_url": "http://localhost:9998/order/list?orderNo=2036116649517522944\u0026paymentNo=P2036116653846044673\u0026paymentResult=cancel",
      "client_reference_id": "P2036116653846044673",
      "client_secret": null,
      "collected_information": null,
      "consent": null,
      "consent_collection": null,
      "created": 1774283021,
      "currency": "cny",
      "currency_conversion": null,
      "custom_fields": [],
      "custom_text": {
        "after_submit": null,
        "shipping_address": null,
        "submit": null,
        "terms_of_service_acceptance": null
      },
      "customer": null,
      "customer_account": null,
      "customer_creation": "if_required",
      "customer_details": {
        "address": {
          "city": null,
          "country": "HK",
          "line1": null,
          "line2": null,
          "postal_code": null,
          "state": null
        },
        "business_name": null,
        "email": "lihua1823494260@gmail.com",
        "individual_name": null,
        "name": "123",
        "phone": null,
        "tax_exempt": "none",
        "tax_ids": []
      },
      "customer_email": null,
      "discounts": [],
      "expires_at": 1774284820,
      "integration_identifier": null,
      "invoice": null,
      "invoice_creation": {
        "enabled": false,
        "invoice_data": {
          "account_tax_ids": null,
          "custom_fields": null,
          "description": null,
          "footer": null,
          "issuer": null,
          "metadata": {},
          "rendering_options": null
        }
      },
      "livemode": false,
      "locale": null,
      "metadata": {
        "user_id": "2035026620841992192",
        "payment_no": "P2036116653846044673",
        "order_no": "2036116649517522944"
      },
      "mode": "payment",
      "origin_context": null,
      "payment_intent": "pi_3TEBGvPm68qiSwnQ01vCMyQq",
      "payment_link": null,
      "payment_method_collection": "if_required",
      "payment_method_configuration_details": null,
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "payment_status": "paid",
      "permissions": null,
      "phone_number_collection": {
        "enabled": false
      },
      "recovered_from": null,
      "saved_payment_method_options": null,
      "setup_intent": null,
      "shipping_address_collection": null,
      "shipping_cost": null,
      "shipping_options": [],
      "status": "complete",
      "submit_type": null,
      "subscription": null,
      "success_url": "http://localhost:9998/payment/processing?orderNo=2036116649517522944\u0026paymentNo=P2036116653846044673\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D",
      "total_details": {
        "amount_discount": 0,
        "amount_shipping": 0,
        "amount_tax": 0
      },
      "ui_mode": "hosted",
      "url": null,
      "wallet_options": null
    }
  },
  "livemode": false,
  "pending_webhooks": 2,
  "request": {
    "id": null,
    "idempotency_key": null
  },
  "type": "checkout.session.completed"
}', '2026-03-24 00:23:34.147', '2026-03-24 00:24:10.923');
INSERT INTO "public"."payment" VALUES (2036126401270128640, 'P2036126401270128641', 2036126401207214080, 2035026620841992192, 1, 560.00, 1, 'cs_test_a1AqtP2vxQndBEaDYeQEPeQLWvO2PZ9sVzdcwW27twJmJE9djLk5gqYw18', '2026-03-24 01:16:32', '{
  "id": "evt_1TEC6BPm68qiSwnQ3BTcwPso",
  "object": "event",
  "api_version": "2025-11-17.clover",
  "created": 1774286227,
  "data": {
    "object": {
      "id": "cs_test_a1AqtP2vxQndBEaDYeQEPeQLWvO2PZ9sVzdcwW27twJmJE9djLk5gqYw18",
      "object": "checkout.session",
      "adaptive_pricing": {
        "enabled": true
      },
      "after_expiration": null,
      "allow_promotion_codes": null,
      "amount_subtotal": 56000,
      "amount_total": 56000,
      "automatic_tax": {
        "enabled": false,
        "liability": null,
        "provider": null,
        "status": null
      },
      "billing_address_collection": null,
      "branding_settings": {
        "background_color": "#ffffff",
        "border_style": "rounded",
        "button_color": "#0074d4",
        "display_name": "New business 沙盒",
        "font_family": "default",
        "icon": null,
        "logo": null
      },
      "cancel_url": "http://localhost:9998/order/list?orderNo=2036126396660588544\u0026paymentNo=P2036126401270128641\u0026paymentResult=cancel",
      "client_reference_id": "P2036126401270128641",
      "client_secret": null,
      "collected_information": null,
      "consent": null,
      "consent_collection": null,
      "created": 1774286192,
      "currency": "cny",
      "currency_conversion": null,
      "custom_fields": [],
      "custom_text": {
        "after_submit": null,
        "shipping_address": null,
        "submit": null,
        "terms_of_service_acceptance": null
      },
      "customer": null,
      "customer_account": null,
      "customer_creation": "if_required",
      "customer_details": {
        "address": {
          "city": null,
          "country": "HK",
          "line1": null,
          "line2": null,
          "postal_code": null,
          "state": null
        },
        "business_name": null,
        "email": "lihua1823494260@gmail.com",
        "individual_name": null,
        "name": "123",
        "phone": null,
        "tax_exempt": "none",
        "tax_ids": []
      },
      "customer_email": null,
      "discounts": [],
      "expires_at": 1774287990,
      "integration_identifier": null,
      "invoice": null,
      "invoice_creation": {
        "enabled": false,
        "invoice_data": {
          "account_tax_ids": null,
          "custom_fields": null,
          "description": null,
          "footer": null,
          "issuer": null,
          "metadata": {},
          "rendering_options": null
        }
      },
      "livemode": false,
      "locale": null,
      "metadata": {
        "user_id": "2035026620841992192",
        "payment_no": "P2036126401270128641",
        "order_no": "2036126396660588544"
      },
      "mode": "payment",
      "origin_context": null,
      "payment_intent": "pi_3TEC6APm68qiSwnQ1BGhak7p",
      "payment_link": null,
      "payment_method_collection": "if_required",
      "payment_method_configuration_details": null,
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "payment_status": "paid",
      "permissions": null,
      "phone_number_collection": {
        "enabled": false
      },
      "recovered_from": null,
      "saved_payment_method_options": null,
      "setup_intent": null,
      "shipping_address_collection": null,
      "shipping_cost": null,
      "shipping_options": [],
      "status": "complete",
      "submit_type": null,
      "subscription": null,
      "success_url": "http://localhost:9998/payment/processing?orderNo=2036126396660588544\u0026paymentNo=P2036126401270128641\u0026paymentResult=success\u0026session_id=%7BCHECKOUT_SESSION_ID%7D",
      "total_details": {
        "amount_discount": 0,
        "amount_shipping": 0,
        "amount_tax": 0
      },
      "ui_mode": "hosted",
      "url": null,
      "wallet_options": null
    }
  },
  "livemode": false,
  "pending_webhooks": 2,
  "request": {
    "id": null,
    "idempotency_key": null
  },
  "type": "checkout.session.completed"
}', '2026-03-24 01:02:18.104', '2026-03-24 01:17:06.846');

-- ----------------------------
-- Table structure for refund
-- ----------------------------
DROP TABLE IF EXISTS "public"."refund";
CREATE TABLE "public"."refund" (
  "id" int8 NOT NULL,
  "refund_no" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "order_id" int8 NOT NULL,
  "payment_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "refund_amount" numeric(12,2) NOT NULL,
  "status" int2 NOT NULL DEFAULT 0,
  "reason" varchar(255) COLLATE "pg_catalog"."default",
  "reject_reason" varchar(255) COLLATE "pg_catalog"."default",
  "trade_no" varchar(128) COLLATE "pg_catalog"."default",
  "audited_by" int8,
  "audited_at" timestamp(3),
  "refunded_at" timestamp(3),
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."refund" IS 'Refund workflow table';

-- ----------------------------
-- Records of refund
-- ----------------------------
INSERT INTO "public"."refund" VALUES (2036116474124312576, 'R2036116474124312577', 2036114045567770624, 2036114045622296576, 2035026620841992192, 1680.00, 3, 'user apply', '', 're_3TEB71Pm68qiSwnQ0lPo5qLQ', 0, '0001-01-01 00:00:00', '2026-03-24 00:22:51.303', '2026-03-24 00:22:51.303', '2026-03-24 00:22:51.303');
INSERT INTO "public"."refund" VALUES (2036116876660056064, 'R2036116876660056065', 2036116653820878848, 2036116653846044672, 2035026620841992192, 1680.00, 3, 'user apply', '', 're_3TEBGvPm68qiSwnQ0d9IXnQW', 0, '0001-01-01 00:00:00', '2026-03-24 00:24:27.275', '2026-03-24 00:24:27.275', '2026-03-24 00:24:27.275');
INSERT INTO "public"."refund" VALUES (2036130198994296832, 'R2036130198994296833', 2036126401207214080, 2036126401270128640, 2035026620841992192, 560.00, 3, 'user apply', '', 're_3TEC6APm68qiSwnQ1ql1u45l', 0, '0001-01-01 00:00:00', '2026-03-24 01:17:23.567', '2026-03-24 01:17:23.567', '2026-03-24 01:17:23.567');

-- ----------------------------
-- Table structure for ticket_buyer
-- ----------------------------
DROP TABLE IF EXISTS "public"."ticket_buyer";
CREATE TABLE "public"."ticket_buyer" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "name" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "id_card" varchar(255) COLLATE "pg_catalog"."default",
  "phone" varchar(20) COLLATE "pg_catalog"."default",
  "is_default" int2 NOT NULL DEFAULT 0,
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."ticket_buyer" IS 'Ticket buyer table';

-- ----------------------------
-- Records of ticket_buyer
-- ----------------------------
INSERT INTO "public"."ticket_buyer" VALUES (2035038227605233664, 2035026620841992192, 'lihua', '7PXPMyukxfw8OkAVu4VA3mIpUTBxz2Fs6+Dw7871U5LXJzVTWVCHl+0ssrAT+w==', '12312341234', 0, '2026-03-21 00:58:17.301', '2026-03-21 01:54:22.577');
INSERT INTO "public"."ticket_buyer" VALUES (2035038076035670016, 2035026620841992192, 'guojian', 'WpPwRzf8KP8rmN6PjvX/JtIdsUKO44h2BfcXytPLz053CTbQV4rDodOmKjYflA==', '18234942601', 1, '2026-03-21 00:57:41.163', '2026-03-21 01:54:22.577');
INSERT INTO "public"."ticket_buyer" VALUES (110000000000000001, 100000000000000001, '陈曦', '310101199201018888', '13800010001', 1, '2025-12-16 09:00:00', '2026-03-20 09:00:00');
INSERT INTO "public"."ticket_buyer" VALUES (110000000000000002, 100000000000000001, '陈雅琴', '310101196905108888', '13800011001', 0, '2025-12-16 09:10:00', '2026-03-20 09:10:00');
INSERT INTO "public"."ticket_buyer" VALUES (110000000000000003, 100000000000000002, '卢露', '310101199408088866', '13800010002', 1, '2025-12-19 10:00:00', '2026-03-20 10:00:00');
INSERT INTO "public"."ticket_buyer" VALUES (110000000000000004, 100000000000000002, '卢建国', '310101197210058877', '13800011002', 0, '2025-12-19 10:05:00', '2026-03-20 10:05:00');
INSERT INTO "public"."ticket_buyer" VALUES (110000000000000005, 100000000000000003, '莫莫', '440101199812128866', '13800010003', 1, '2026-01-04 11:20:00', '2026-03-20 11:20:00');
INSERT INTO "public"."ticket_buyer" VALUES (110000000000000006, 100000000000000004, '赵凯文', '440101199305056666', '13800010004', 1, '2026-01-11 14:30:00', '2026-03-20 14:30:00');
INSERT INTO "public"."ticket_buyer" VALUES (110000000000000007, 100000000000000005, '李文雯', '510101199611126666', '13800010005', 1, '2026-01-17 15:30:00', '2026-03-20 15:30:00');
INSERT INTO "public"."ticket_buyer" VALUES (110000000000000008, 100000000000000006, '孙娜娜', '330101199902018899', '13800010006', 1, '2026-02-03 16:30:00', '2026-03-20 16:30:00');

-- ----------------------------
-- Table structure for ticket_tier
-- ----------------------------
DROP TABLE IF EXISTS "public"."ticket_tier";
CREATE TABLE "public"."ticket_tier" (
  "id" int8 NOT NULL,
  "event_id" int8 NOT NULL,
  "name" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "price" numeric(10,2) NOT NULL,
  "total_stock" int4 NOT NULL DEFAULT 0,
  "sold_count" int4 NOT NULL DEFAULT 0,
  "locked_count" int4 NOT NULL DEFAULT 0,
  "status" int2 NOT NULL DEFAULT 1,
  "sort_order" int4 NOT NULL DEFAULT 0,
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."ticket_tier" IS 'Ticket tier and inventory table';

-- ----------------------------
-- Records of ticket_tier
-- ----------------------------
INSERT INTO "public"."ticket_tier" VALUES (150000000000000002, 140000000000000001, '看台 A 档', 980.00, 5000, 3200, 120, 1, 20, '2026-02-20 10:15:00', '2026-03-20 12:00:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000003, 140000000000000001, '看台 B 档', 680.00, 6000, 3900, 160, 1, 30, '2026-02-20 10:20:00', '2026-03-20 12:00:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000005, 140000000000000002, '看台 A 区', 880.00, 4500, 2800, 90, 1, 20, '2026-02-25 11:15:00', '2026-03-20 12:05:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000006, 140000000000000002, '看台 B 区', 580.00, 5000, 3400, 110, 1, 30, '2026-02-25 11:20:00', '2026-03-20 12:05:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000007, 140000000000000003, '早鸟单日票', 399.00, 3000, 3000, 0, 2, 10, '2026-02-28 14:10:00', '2026-03-20 12:10:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000009, 140000000000000003, '两日通票', 899.00, 3500, 1800, 120, 1, 30, '2026-02-28 14:20:00', '2026-03-20 12:10:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000010, 140000000000000004, 'VIP', 1080.00, 500, 220, 8, 1, 10, '2026-03-01 09:40:00', '2026-03-20 12:15:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000012, 140000000000000004, '二等座', 380.00, 1800, 910, 20, 1, 30, '2026-03-01 09:50:00', '2026-03-20 12:15:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000013, 140000000000000005, '场边 VIP', 2280.00, 200, 200, 0, 2, 10, '2026-03-03 11:10:00', '2026-03-20 12:20:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000015, 140000000000000005, 'B 档看台', 680.00, 2800, 1850, 60, 1, 30, '2026-03-03 11:20:00', '2026-03-20 12:20:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000016, 140000000000000006, '工作日票', 98.00, 20000, 6200, 120, 1, 10, '2026-03-05 13:10:00', '2026-03-20 12:25:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000017, 140000000000000006, '周末票', 128.00, 12000, 5100, 90, 1, 20, '2026-03-05 13:15:00', '2026-03-20 12:25:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000018, 140000000000000006, '亲子套票', 228.00, 5000, 2400, 40, 1, 30, '2026-03-05 13:20:00', '2026-03-20 12:25:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000019, 140000000000000007, 'VIP 亲子套票', 680.00, 400, 210, 5, 1, 10, '2026-03-07 09:10:00', '2026-03-20 12:30:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000020, 140000000000000007, '普通票', 280.00, 1500, 760, 22, 1, 20, '2026-03-07 09:15:00', '2026-03-20 12:30:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000022, 140000000000000008, '看台 A 区', 1180.00, 4800, 1880, 55, 1, 20, '2026-03-09 12:15:00', '2026-03-20 12:35:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000025, 140000000000000009, '一等看台', 980.00, 4200, 1660, 35, 1, 20, '2026-03-12 10:15:00', '2026-03-20 12:40:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000026, 140000000000000009, '二等看台', 580.00, 5800, 2380, 50, 1, 30, '2026-03-12 10:20:00', '2026-03-20 12:40:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000028, 140000000000000010, '标准席', 480.00, 1000, 260, 10, 1, 20, '2026-03-14 16:15:00', '2026-03-20 12:45:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000030, 140000000000000011, '内场', 1550.00, 800, 530, 0, 0, 10, '2026-03-15 09:10:00', '2026-03-20 12:50:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000031, 140000000000000011, '看台', 850.00, 4200, 2400, 0, 0, 20, '2026-03-15 09:15:00', '2026-03-20 12:50:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000032, 140000000000000012, '欢乐跑', 199.00, 8000, 7900, 0, 1, 10, '2025-08-01 09:10:00', '2025-11-09 12:35:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000033, 140000000000000012, '半程组', 399.00, 3000, 3000, 0, 2, 20, '2025-08-01 09:15:00', '2025-11-09 12:35:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000034, 140000000000000013, 'VIP', 480.00, 300, 0, 0, 0, 10, '2026-03-16 11:10:00', '2026-03-16 11:10:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000035, 140000000000000013, '普通票', 180.00, 1000, 0, 0, 0, 20, '2026-03-16 11:15:00', '2026-03-16 11:15:00');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000011, 140000000000000004, '一等座', 680.00, 1200, 730, 13, 1, 20, '2026-03-01 09:45:00', '2026-03-23 09:26:41.677');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000014, 140000000000000005, 'A 档看台', 1280.00, 1600, 1240, 30, 1, 20, '2026-03-03 11:15:00', '2026-03-23 09:46:39.976');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000027, 140000000000000010, '特邀席', 880.00, 300, 80, 4, 1, 10, '2026-03-14 16:10:00', '2026-03-23 10:02:43.287');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000008, 140000000000000003, '预售单日票', 499.00, 6000, 4200, 201, 1, 20, '2026-02-28 14:15:00', '2026-03-23 10:11:09.934');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000001, 140000000000000001, '内场 VIP', 1880.00, 1200, 862, 40, 1, 10, '2026-02-20 10:10:00', '2026-03-23 23:05:46.503');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000021, 140000000000000008, '内场', 1980.00, 900, 521, 18, 1, 10, '2026-03-09 12:10:00', '2026-03-23 23:29:31.182');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000023, 140000000000000008, '看台 B 区', 780.00, 5200, 2141, 70, 1, 30, '2026-03-09 12:20:00', '2026-03-23 23:46:46.361');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000024, 140000000000000009, '贵宾票', 1680.00, 600, 240, 10, 1, 10, '2026-03-12 10:10:00', '2026-03-24 00:22:51.299');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000004, 140000000000000002, '内场', 1680.00, 1000, 721, 30, 1, 10, '2026-02-25 11:10:00', '2026-03-24 00:24:27.27');
INSERT INTO "public"."ticket_tier" VALUES (150000000000000029, 140000000000000010, '青年票', 280.00, 800, 150, 6, 1, 30, '2026-03-14 16:20:00', '2026-03-24 01:17:23.56');

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS "public"."user";
CREATE TABLE "public"."user" (
  "id" int8 NOT NULL,
  "username" varchar(64) COLLATE "pg_catalog"."default",
  "password_hash" varchar(255) COLLATE "pg_catalog"."default",
  "phone" varchar(20) COLLATE "pg_catalog"."default",
  "email" varchar(128) COLLATE "pg_catalog"."default",
  "nickname" varchar(64) COLLATE "pg_catalog"."default",
  "avatar" varchar(255) COLLATE "pg_catalog"."default",
  "status" int2 NOT NULL DEFAULT 1,
  "is_verified" int2 NOT NULL DEFAULT 0,
  "real_name" varchar(255) COLLATE "pg_catalog"."default",
  "id_card" varchar(255) COLLATE "pg_catalog"."default",
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamp(3)
)
;
COMMENT ON TABLE "public"."user" IS 'Core user account table';

-- ----------------------------
-- Records of user
-- ----------------------------
INSERT INTO "public"."user" VALUES (2035026620841992192, '', '$2a$10$gBPuFiNfIO2fYh44OZf.r.Jte9N/3ezskkzki/CovL0rMxfvo/Cjm', '18303802780', '', '183****2780', 'https://avatars.githubusercontent.com/u/6800565?s=48&v=4', 1, 0, '', '', '2026-03-21 00:12:10.033', '2026-03-21 01:55:02.378', NULL);
INSERT INTO "public"."user" VALUES (100000000000000001, 'chenxi', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '13800010001', 'chenxi@example.com', '晨曦', 'https://cdn.example.com/avatar/chenxi.jpg', 1, 1, '陈曦', '310101199201018888', '2025-12-15 09:30:00', '2026-03-20 09:30:00', NULL);
INSERT INTO "public"."user" VALUES (100000000000000002, 'lulu', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '13800010002', 'lulu@example.com', '露露', 'https://cdn.example.com/avatar/lulu.jpg', 1, 1, '卢露', '310101199408088866', '2025-12-18 10:15:00', '2026-03-20 10:15:00', NULL);
INSERT INTO "public"."user" VALUES (100000000000000003, 'momo', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '13800010003', 'momo@example.com', '莫莫', 'https://cdn.example.com/avatar/momo.jpg', 1, 0, NULL, NULL, '2026-01-03 12:00:00', '2026-03-20 12:00:00', NULL);
INSERT INTO "public"."user" VALUES (100000000000000004, 'kevin', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '13800010004', 'kevin@example.com', 'Kevin', 'https://cdn.example.com/avatar/kevin.jpg', 1, 1, '赵凯文', '440101199305056666', '2026-01-10 15:20:00', '2026-03-20 15:20:00', NULL);
INSERT INTO "public"."user" VALUES (100000000000000005, 'wenwen', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '13800010005', 'wenwen@example.com', '文雯', 'https://cdn.example.com/avatar/wenwen.jpg', 1, 1, '李文雯', '510101199611126666', '2026-01-16 16:40:00', '2026-03-20 16:40:00', NULL);
INSERT INTO "public"."user" VALUES (100000000000000006, 'nana', '$2a$10$DgTmBWeMT04t6okSXdPJh.SV05mVjJXU7ME.mFHkqxE63BHjIul.m', '13800010006', 'nana@example.com', '娜娜', 'https://cdn.example.com/avatar/nana.jpg', 1, 0, NULL, NULL, '2026-02-02 18:10:00', '2026-03-20 18:10:00', NULL);

-- ----------------------------
-- Table structure for venue
-- ----------------------------
DROP TABLE IF EXISTS "public"."venue";
CREATE TABLE "public"."venue" (
  "id" int8 NOT NULL,
  "name" varchar(128) COLLATE "pg_catalog"."default" NOT NULL,
  "city_id" int8 NOT NULL,
  "address" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "capacity" int4,
  "seat_map_url" varchar(255) COLLATE "pg_catalog"."default",
  "description" text COLLATE "pg_catalog"."default",
  "created_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
COMMENT ON TABLE "public"."venue" IS 'Venue table';

-- ----------------------------
-- Records of venue
-- ----------------------------
INSERT INTO "public"."venue" VALUES (130000000000000002, '上海梅赛德斯奔驰文化中心', 2, '上海市浦东新区世博大道1200号', 18000, 'https://cdn.example.com/seatmap/venue-mercedes.png', '华东地区热门室内演出场馆，舞台设备成熟。', '2025-12-01 08:10:00', '2026-03-20 08:10:00');
INSERT INTO "public"."venue" VALUES (130000000000000003, '成都露天音乐公园', 1, '成都市金牛区北星大道一段', 30000, 'https://cdn.example.com/seatmap/venue-cdpark.png', '适合音乐节与大型户外演出，氛围感强。', '2025-12-01 08:20:00', '2026-03-20 08:20:00');
INSERT INTO "public"."venue" VALUES (130000000000000004, '杭州大剧院', 1, '杭州市上城区新业路39号', 1800, 'https://cdn.example.com/seatmap/venue-hzdj.png', '以舞剧、歌剧和高品质舞台演出为主。', '2025-12-01 08:30:00', '2026-03-20 08:30:00');
INSERT INTO "public"."venue" VALUES (130000000000000005, '广州国际体育演艺中心', 1, '广州市黄埔区开创大道2666号', 16000, 'https://cdn.example.com/seatmap/venue-gzarena.png', '华南大型文体场馆，适配演唱会和篮球赛事。', '2025-12-01 08:40:00', '2026-03-20 08:40:00');
INSERT INTO "public"."venue" VALUES (130000000000000006, '深圳湾体育中心春茧', 3, '深圳市南山区滨海大道3001号', 20000, 'https://cdn.example.com/seatmap/venue-szbay.png', '深圳热门综合场馆，交通便利。', '2025-12-01 08:50:00', '2026-03-20 08:50:00');
INSERT INTO "public"."venue" VALUES (130000000000000007, '上海世博展览馆', 2, '上海市浦东新区国展路1099号', 12000, 'https://cdn.example.com/seatmap/venue-expo.png', '适合沉浸展与大型主题展览。', '2025-12-01 09:00:00', '2026-03-20 09:00:00');
INSERT INTO "public"."venue" VALUES (130000000000000008, '南京青奥体育公园体育馆', 1, '南京市建邺区江山大街8号', 21000, 'https://cdn.example.com/seatmap/venue-njolympic.png', '适合大型巡演与篮球比赛。', '2025-12-01 09:10:00', '2026-03-20 09:10:00');
INSERT INTO "public"."venue" VALUES (130000000000000009, '西安曲江国际会展中心', 4, '西安市雁塔区汇新路15号', 8000, 'https://cdn.example.com/seatmap/venue-xianexpo.png', '适合儿童演出、展览和动漫活动。', '2025-12-01 09:20:00', '2026-03-20 09:20:00');
INSERT INTO "public"."venue" VALUES (130000000000000010, '北京天桥艺术中心', 1, '北京市西城区天桥南大街9号', 2200, 'https://cdn.example.com/seatmap/venue-bjtheatre.png', '适合舞剧、音乐剧和中型戏剧演出。', '2025-12-01 09:30:00', '2026-03-20 09:30:00');
INSERT INTO "public"."venue" VALUES (130000000000000001, '北京国家体育场（鸟巢）', 1, '北京市朝阳区国家体育场南路1号', 91000, 'https://img.alicdn.com/bao/uploaded/i2/2251059038/O1CN01E3LfNG2GdSlslh9hu_!!4611686018427383646-0-item_pic.jpg_q60.jpg_.webp', '大型综合体育场，适合超大型演唱会与体育赛事。', '2025-12-01 08:00:00', '2026-03-21 19:11:19.289');

-- ----------------------------
-- Function structure for set_updated_at
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."set_updated_at"();
CREATE OR REPLACE FUNCTION "public"."set_updated_at"()
  RETURNS "pg_catalog"."trigger" AS $BODY$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."category_id_seq"
OWNED BY "public"."category"."id";
SELECT setval('"public"."category_id_seq"', 6, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."category_id_seq1"
OWNED BY "public"."category"."id";
SELECT setval('"public"."category_id_seq1"', 1, false);

-- ----------------------------
-- Indexes structure for table address
-- ----------------------------
CREATE INDEX "idx_address_default" ON "public"."address" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "is_default" "pg_catalog"."int2_ops" ASC NULLS LAST
);
CREATE INDEX "idx_address_user" ON "public"."address" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table address
-- ----------------------------
CREATE TRIGGER "trg_address_set_updated_at" BEFORE UPDATE ON "public"."address"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Checks structure for table address
-- ----------------------------
ALTER TABLE "public"."address" ADD CONSTRAINT "chk_address_is_default" CHECK (is_default = ANY (ARRAY[0, 1]));

-- ----------------------------
-- Primary Key structure for table address
-- ----------------------------
ALTER TABLE "public"."address" ADD CONSTRAINT "address_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table admin_log
-- ----------------------------
CREATE INDEX "idx_admin_log_admin_time" ON "public"."admin_log" USING btree (
  "admin_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_admin_log_target" ON "public"."admin_log" USING btree (
  "target_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "target_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table admin_log
-- ----------------------------
ALTER TABLE "public"."admin_log" ADD CONSTRAINT "admin_log_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table admin_user
-- ----------------------------
CREATE INDEX "idx_admin_user_status" ON "public"."admin_user" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table admin_user
-- ----------------------------
CREATE TRIGGER "trg_admin_user_set_updated_at" BEFORE UPDATE ON "public"."admin_user"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Uniques structure for table admin_user
-- ----------------------------
ALTER TABLE "public"."admin_user" ADD CONSTRAINT "uk_admin_user_username" UNIQUE ("username");
ALTER TABLE "public"."admin_user" ADD CONSTRAINT "uk_admin_user_phone" UNIQUE ("phone");

-- ----------------------------
-- Checks structure for table admin_user
-- ----------------------------
ALTER TABLE "public"."admin_user" ADD CONSTRAINT "chk_admin_user_role" CHECK (role = ANY (ARRAY[1, 2, 3]));
ALTER TABLE "public"."admin_user" ADD CONSTRAINT "chk_admin_user_status" CHECK (status = ANY (ARRAY[0, 1]));

-- ----------------------------
-- Primary Key structure for table admin_user
-- ----------------------------
ALTER TABLE "public"."admin_user" ADD CONSTRAINT "admin_user_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for category
-- ----------------------------
SELECT setval('"public"."category_id_seq1"', 1, false);

-- ----------------------------
-- Indexes structure for table category
-- ----------------------------
CREATE INDEX "idx_category_status_sort" ON "public"."category" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."int4_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table category
-- ----------------------------
ALTER TABLE "public"."category" ADD CONSTRAINT "uk_category_name" UNIQUE ("name");

-- ----------------------------
-- Checks structure for table category
-- ----------------------------
ALTER TABLE "public"."category" ADD CONSTRAINT "chk_category_status" CHECK (status = ANY (ARRAY[0, 1]));

-- ----------------------------
-- Primary Key structure for table category
-- ----------------------------
ALTER TABLE "public"."category" ADD CONSTRAINT "category_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table city
-- ----------------------------
ALTER TABLE "public"."city" ADD CONSTRAINT "city_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table event
-- ----------------------------
CREATE INDEX "idx_event_category_status" ON "public"."event" USING btree (
  "category_id" "pg_catalog"."int4_ops" ASC NULLS LAST,
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST
);
CREATE INDEX "idx_event_city_start" ON "public"."event" USING btree (
  "city_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "event_start_time" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_event_created_by" ON "public"."event" USING btree (
  "created_by" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_event_deleted_at" ON "public"."event" USING btree (
  "deleted_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_event_sale_time" ON "public"."event" USING btree (
  "sale_start_time" "pg_catalog"."timestamp_ops" ASC NULLS LAST,
  "sale_end_time" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_event_venue" ON "public"."event" USING btree (
  "venue_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table event
-- ----------------------------
CREATE TRIGGER "trg_event_set_updated_at" BEFORE UPDATE ON "public"."event"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Checks structure for table event
-- ----------------------------
ALTER TABLE "public"."event" ADD CONSTRAINT "chk_event_need_real_name" CHECK (need_real_name = ANY (ARRAY[0, 1]));
ALTER TABLE "public"."event" ADD CONSTRAINT "chk_event_ticket_type" CHECK (ticket_type = ANY (ARRAY[1, 2]));
ALTER TABLE "public"."event" ADD CONSTRAINT "chk_event_time_window" CHECK (event_end_time >= event_start_time);
ALTER TABLE "public"."event" ADD CONSTRAINT "chk_event_sale_window" CHECK (sale_end_time >= sale_start_time);
ALTER TABLE "public"."event" ADD CONSTRAINT "chk_event_status" CHECK (status = ANY (ARRAY[0, 1, 2, 3]));
ALTER TABLE "public"."event" ADD CONSTRAINT "chk_event_purchase_limit" CHECK (purchase_limit >= 1);

-- ----------------------------
-- Primary Key structure for table event
-- ----------------------------
ALTER TABLE "public"."event" ADD CONSTRAINT "event_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table event_favorite
-- ----------------------------
CREATE INDEX "idx_event_favorite_event" ON "public"."event_favorite" USING btree (
  "event_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table event_favorite
-- ----------------------------
ALTER TABLE "public"."event_favorite" ADD CONSTRAINT "uk_event_favorite_user_event" UNIQUE ("user_id", "event_id");

-- ----------------------------
-- Checks structure for table event_favorite
-- ----------------------------
ALTER TABLE "public"."event_favorite" ADD CONSTRAINT "chk_event_favorite_notify_enabled" CHECK (notify_enabled = ANY (ARRAY[0, 1]));

-- ----------------------------
-- Primary Key structure for table event_favorite
-- ----------------------------
ALTER TABLE "public"."event_favorite" ADD CONSTRAINT "event_favorite_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table notification
-- ----------------------------
CREATE INDEX "idx_notification_type_time" ON "public"."notification" USING btree (
  "type" "pg_catalog"."int2_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_notification_user_read_time" ON "public"."notification" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "is_read" "pg_catalog"."int2_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);

-- ----------------------------
-- Checks structure for table notification
-- ----------------------------
ALTER TABLE "public"."notification" ADD CONSTRAINT "chk_notification_type" CHECK (type = ANY (ARRAY[1, 2, 3, 4]));
ALTER TABLE "public"."notification" ADD CONSTRAINT "chk_notification_is_read" CHECK (is_read = ANY (ARRAY[0, 1]));

-- ----------------------------
-- Primary Key structure for table notification
-- ----------------------------
ALTER TABLE "public"."notification" ADD CONSTRAINT "notification_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table order_info
-- ----------------------------
CREATE INDEX "idx_order_info_address" ON "public"."order_info" USING btree (
  "address_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_order_info_event" ON "public"."order_info" USING btree (
  "event_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_order_info_pay_deadline" ON "public"."order_info" USING btree (
  "pay_deadline" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_order_info_status_created" ON "public"."order_info" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_order_info_ticket_tier" ON "public"."order_info" USING btree (
  "ticket_tier_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_order_info_user_created" ON "public"."order_info" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table order_info
-- ----------------------------
CREATE TRIGGER "trg_order_info_set_updated_at" BEFORE UPDATE ON "public"."order_info"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Uniques structure for table order_info
-- ----------------------------
ALTER TABLE "public"."order_info" ADD CONSTRAINT "uk_order_info_order_no" UNIQUE ("order_no");

-- ----------------------------
-- Checks structure for table order_info
-- ----------------------------
ALTER TABLE "public"."order_info" ADD CONSTRAINT "chk_order_info_unit_price" CHECK (unit_price >= 0::numeric);
ALTER TABLE "public"."order_info" ADD CONSTRAINT "chk_order_info_status" CHECK (status = ANY (ARRAY[0, 1, 2, 3, 4, 5]));
ALTER TABLE "public"."order_info" ADD CONSTRAINT "chk_order_info_cancel_reason" CHECK (cancel_reason = ANY (ARRAY[0, 1, 2, 3]));
ALTER TABLE "public"."order_info" ADD CONSTRAINT "chk_order_info_total_amount" CHECK (total_amount >= 0::numeric);
ALTER TABLE "public"."order_info" ADD CONSTRAINT "chk_order_info_quantity" CHECK (quantity >= 1);

-- ----------------------------
-- Primary Key structure for table order_info
-- ----------------------------
ALTER TABLE "public"."order_info" ADD CONSTRAINT "order_info_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table order_ticket
-- ----------------------------
CREATE INDEX "idx_order_ticket_buyer" ON "public"."order_ticket" USING btree (
  "ticket_buyer_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_order_ticket_order" ON "public"."order_ticket" USING btree (
  "order_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_order_ticket_status" ON "public"."order_ticket" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table order_ticket
-- ----------------------------
CREATE TRIGGER "trg_order_ticket_set_updated_at" BEFORE UPDATE ON "public"."order_ticket"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Uniques structure for table order_ticket
-- ----------------------------
ALTER TABLE "public"."order_ticket" ADD CONSTRAINT "uk_order_ticket_ticket_code" UNIQUE ("ticket_code");

-- ----------------------------
-- Checks structure for table order_ticket
-- ----------------------------
ALTER TABLE "public"."order_ticket" ADD CONSTRAINT "chk_order_ticket_status" CHECK (status = ANY (ARRAY[0, 1, 2]));

-- ----------------------------
-- Primary Key structure for table order_ticket
-- ----------------------------
ALTER TABLE "public"."order_ticket" ADD CONSTRAINT "order_ticket_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table payment
-- ----------------------------
CREATE INDEX "idx_payment_order_id" ON "public"."payment" USING btree (
  "order_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_payment_status_created" ON "public"."payment" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_payment_user_created" ON "public"."payment" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table payment
-- ----------------------------
CREATE TRIGGER "trg_payment_set_updated_at" BEFORE UPDATE ON "public"."payment"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Uniques structure for table payment
-- ----------------------------
ALTER TABLE "public"."payment" ADD CONSTRAINT "uk_payment_payment_no" UNIQUE ("payment_no");
ALTER TABLE "public"."payment" ADD CONSTRAINT "uk_payment_trade_no" UNIQUE ("trade_no");

-- ----------------------------
-- Checks structure for table payment
-- ----------------------------
ALTER TABLE "public"."payment" ADD CONSTRAINT "chk_payment_amount" CHECK (amount >= 0::numeric);
ALTER TABLE "public"."payment" ADD CONSTRAINT "chk_payment_pay_method" CHECK (pay_method = ANY (ARRAY[1, 2]));
ALTER TABLE "public"."payment" ADD CONSTRAINT "chk_payment_status" CHECK (status = ANY (ARRAY[0, 1, 2]));

-- ----------------------------
-- Primary Key structure for table payment
-- ----------------------------
ALTER TABLE "public"."payment" ADD CONSTRAINT "payment_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table refund
-- ----------------------------
CREATE INDEX "idx_refund_audited_by" ON "public"."refund" USING btree (
  "audited_by" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_refund_status_created" ON "public"."refund" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_refund_user_created" ON "public"."refund" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table refund
-- ----------------------------
CREATE TRIGGER "trg_refund_set_updated_at" BEFORE UPDATE ON "public"."refund"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Uniques structure for table refund
-- ----------------------------
ALTER TABLE "public"."refund" ADD CONSTRAINT "uk_refund_refund_no" UNIQUE ("refund_no");
ALTER TABLE "public"."refund" ADD CONSTRAINT "uk_refund_order_id" UNIQUE ("order_id");
ALTER TABLE "public"."refund" ADD CONSTRAINT "uk_refund_payment_id" UNIQUE ("payment_id");
ALTER TABLE "public"."refund" ADD CONSTRAINT "uk_refund_trade_no" UNIQUE ("trade_no");

-- ----------------------------
-- Checks structure for table refund
-- ----------------------------
ALTER TABLE "public"."refund" ADD CONSTRAINT "chk_refund_status" CHECK (status = ANY (ARRAY[0, 1, 2, 3, 4, 5]));
ALTER TABLE "public"."refund" ADD CONSTRAINT "chk_refund_amount" CHECK (refund_amount >= 0::numeric);

-- ----------------------------
-- Primary Key structure for table refund
-- ----------------------------
ALTER TABLE "public"."refund" ADD CONSTRAINT "refund_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table ticket_buyer
-- ----------------------------
CREATE INDEX "idx_ticket_buyer_default" ON "public"."ticket_buyer" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "is_default" "pg_catalog"."int2_ops" ASC NULLS LAST
);
CREATE INDEX "idx_ticket_buyer_user" ON "public"."ticket_buyer" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table ticket_buyer
-- ----------------------------
CREATE TRIGGER "trg_ticket_buyer_set_updated_at" BEFORE UPDATE ON "public"."ticket_buyer"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Checks structure for table ticket_buyer
-- ----------------------------
ALTER TABLE "public"."ticket_buyer" ADD CONSTRAINT "chk_ticket_buyer_is_default" CHECK (is_default = ANY (ARRAY[0, 1]));

-- ----------------------------
-- Primary Key structure for table ticket_buyer
-- ----------------------------
ALTER TABLE "public"."ticket_buyer" ADD CONSTRAINT "ticket_buyer_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table ticket_tier
-- ----------------------------
CREATE INDEX "idx_ticket_tier_event_status" ON "public"."ticket_tier" USING btree (
  "event_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table ticket_tier
-- ----------------------------
CREATE TRIGGER "trg_ticket_tier_set_updated_at" BEFORE UPDATE ON "public"."ticket_tier"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Uniques structure for table ticket_tier
-- ----------------------------
ALTER TABLE "public"."ticket_tier" ADD CONSTRAINT "uk_ticket_tier_event_name" UNIQUE ("event_id", "name");

-- ----------------------------
-- Checks structure for table ticket_tier
-- ----------------------------
ALTER TABLE "public"."ticket_tier" ADD CONSTRAINT "chk_ticket_tier_price" CHECK (price >= 0::numeric);
ALTER TABLE "public"."ticket_tier" ADD CONSTRAINT "chk_ticket_tier_status" CHECK (status = ANY (ARRAY[0, 1, 2]));
ALTER TABLE "public"."ticket_tier" ADD CONSTRAINT "chk_ticket_tier_stock" CHECK (total_stock >= (sold_count + locked_count));

-- ----------------------------
-- Primary Key structure for table ticket_tier
-- ----------------------------
ALTER TABLE "public"."ticket_tier" ADD CONSTRAINT "ticket_tier_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table user
-- ----------------------------
CREATE INDEX "idx_user_deleted_at" ON "public"."user" USING btree (
  "deleted_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_user_email" ON "public"."user" USING btree (
  "email" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_user_status" ON "public"."user" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table user
-- ----------------------------
CREATE TRIGGER "trg_user_set_updated_at" BEFORE UPDATE ON "public"."user"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Uniques structure for table user
-- ----------------------------
ALTER TABLE "public"."user" ADD CONSTRAINT "uk_user_phone" UNIQUE ("phone");

-- ----------------------------
-- Checks structure for table user
-- ----------------------------
ALTER TABLE "public"."user" ADD CONSTRAINT "chk_user_status" CHECK (status = ANY (ARRAY[0, 1]));
ALTER TABLE "public"."user" ADD CONSTRAINT "chk_user_is_verified" CHECK (is_verified = ANY (ARRAY[0, 1]));

-- ----------------------------
-- Primary Key structure for table user
-- ----------------------------
ALTER TABLE "public"."user" ADD CONSTRAINT "user_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table venue
-- ----------------------------
CREATE INDEX "idx_venue_city" ON "public"."venue" USING btree (
  "city_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table venue
-- ----------------------------
CREATE TRIGGER "trg_venue_set_updated_at" BEFORE UPDATE ON "public"."venue"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Checks structure for table venue
-- ----------------------------
ALTER TABLE "public"."venue" ADD CONSTRAINT "chk_venue_capacity" CHECK (capacity IS NULL OR capacity >= 0);

-- ----------------------------
-- Primary Key structure for table venue
-- ----------------------------
ALTER TABLE "public"."venue" ADD CONSTRAINT "venue_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Foreign Keys structure for table address
-- ----------------------------
ALTER TABLE "public"."address" ADD CONSTRAINT "fk_address_user" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table admin_log
-- ----------------------------
ALTER TABLE "public"."admin_log" ADD CONSTRAINT "fk_admin_log_admin_user" FOREIGN KEY ("admin_id") REFERENCES "public"."admin_user" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table event
-- ----------------------------
ALTER TABLE "public"."event" ADD CONSTRAINT "fk_event_category" FOREIGN KEY ("category_id") REFERENCES "public"."category" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "public"."event" ADD CONSTRAINT "fk_event_created_by_admin" FOREIGN KEY ("created_by") REFERENCES "public"."admin_user" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "public"."event" ADD CONSTRAINT "fk_event_venue" FOREIGN KEY ("venue_id") REFERENCES "public"."venue" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table event_favorite
-- ----------------------------
ALTER TABLE "public"."event_favorite" ADD CONSTRAINT "fk_event_favorite_event" FOREIGN KEY ("event_id") REFERENCES "public"."event" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "public"."event_favorite" ADD CONSTRAINT "fk_event_favorite_user" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table notification
-- ----------------------------
ALTER TABLE "public"."notification" ADD CONSTRAINT "fk_notification_user" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table order_ticket
-- ----------------------------
ALTER TABLE "public"."order_ticket" ADD CONSTRAINT "fk_order_ticket_order" FOREIGN KEY ("order_id") REFERENCES "public"."order_info" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "public"."order_ticket" ADD CONSTRAINT "fk_order_ticket_ticket_buyer" FOREIGN KEY ("ticket_buyer_id") REFERENCES "public"."ticket_buyer" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table payment
-- ----------------------------
ALTER TABLE "public"."payment" ADD CONSTRAINT "fk_payment_order_info" FOREIGN KEY ("order_id") REFERENCES "public"."order_info" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "public"."payment" ADD CONSTRAINT "fk_payment_user" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table refund
-- ----------------------------
ALTER TABLE "public"."refund" ADD CONSTRAINT "fk_refund_order_info" FOREIGN KEY ("order_id") REFERENCES "public"."order_info" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "public"."refund" ADD CONSTRAINT "fk_refund_payment" FOREIGN KEY ("payment_id") REFERENCES "public"."payment" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "public"."refund" ADD CONSTRAINT "fk_refund_user" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table ticket_buyer
-- ----------------------------
ALTER TABLE "public"."ticket_buyer" ADD CONSTRAINT "fk_ticket_buyer_user" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- ----------------------------
-- Foreign Keys structure for table ticket_tier
-- ----------------------------
ALTER TABLE "public"."ticket_tier" ADD CONSTRAINT "fk_ticket_tier_event" FOREIGN KEY ("event_id") REFERENCES "public"."event" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
