-- 演示数据（可重复执行：依赖 slug / sku 唯一性时按需调整）
INSERT INTO categories (name, slug, description, sort_order, is_hot, status)
VALUES
    ('芝士奶盖', 'cheese-foam', '奶盖类', 100, TRUE, 1),
    ('鲜果茶', 'fruit-tea', '鲜果与茶底', 90, TRUE, 1),
    ('经典奶茶', 'classic-milk-tea', '传统奶茶', 80, FALSE, 1)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO teas (category_id, name, shop_name, subtitle, description, image_url, tags, price_min_cent, price_max_cent, sugar_level, is_recommended, status)
SELECT c.id, '芝芝莓莓', '喜茶', '草莓季限定', '草莓果肉 + 芝士奶盖', '', ARRAY['芝士','草莓'], 1900, 2600, '少糖', TRUE, 1
FROM categories c WHERE c.slug = 'cheese-foam'
AND NOT EXISTS (SELECT 1 FROM teas t WHERE t.name = '芝芝莓莓' AND t.shop_name = '喜茶');

-- teas 无 slug 唯一约束，重复执行会插入多条；生产环境请用迁移工具或幂等脚本
-- 下面用 NOT EXISTS 保证演示数据只插一次
INSERT INTO teas (category_id, name, shop_name, subtitle, description, image_url, tags, price_min_cent, price_max_cent, sugar_level, is_recommended, status)
SELECT c.id, '杨枝甘露', '七分甜', '芒果椰奶', '经典港式风味', '', ARRAY['芒果','椰奶'], 1600, 2200, '标准', TRUE, 1
FROM categories c WHERE c.slug = 'fruit-tea'
AND NOT EXISTS (SELECT 1 FROM teas t WHERE t.name = '杨枝甘露' AND t.shop_name = '七分甜');

INSERT INTO teas (category_id, name, shop_name, subtitle, description, image_url, tags, price_min_cent, price_max_cent, sugar_level, is_recommended, status)
SELECT c.id, '珍珠奶茶', '一点点', '经典款', '黑糖珍珠 + 奶茶底', '', ARRAY['珍珠','经典'], 1200, 1800, '半糖', FALSE, 1
FROM categories c WHERE c.slug = 'classic-milk-tea'
AND NOT EXISTS (SELECT 1 FROM teas t WHERE t.name = '珍珠奶茶' AND t.shop_name = '一点点');
