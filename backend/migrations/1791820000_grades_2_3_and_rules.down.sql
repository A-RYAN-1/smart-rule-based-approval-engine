DELETE FROM rules WHERE grade_id IN (SELECT id FROM grades WHERE name IN ('Grade 2', 'Grade 3'));
DELETE FROM grades WHERE name IN ('Grade 2', 'Grade 3');
