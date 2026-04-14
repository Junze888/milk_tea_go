-- 演示数据回滚（按名称删除，生产请谨慎）
DELETE FROM teas WHERE name IN ('芝芝莓莓', '杨枝甘露', '珍珠奶茶');
DELETE FROM categories WHERE slug IN ('cheese-foam', 'fruit-tea', 'classic-milk-tea');
