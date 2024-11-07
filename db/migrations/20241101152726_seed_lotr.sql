-- +goose Up
-- +goose StatementBegin
INSERT INTO facilities (name, code) VALUES
    ('Rivendell Medical Center', 'RVDL'),
    ('Minas Tirith General Hospital', 'MTGH'),
    ('Lothlorien Health Center', 'LOTH'),
    ('Erebor Mountain Clinic', 'ERBR'),
    ('Shire Community Hospital', 'SHIR')
ON CONFLICT (code) DO NOTHING;

-- Then seed controllers (using facilities' IDs)
INSERT INTO controllers (first_name, last_name, initials, email, password, facility_id, role) VALUES
    -- Rivendell Staff
    ('Logan', 'Williams', 'LW', 'logan@fireflysoftware.dev', '$2a$10$R40Petx4mfIZQHen70VoleqZ6B.IxtKewvW1eS.WgXYQmiS/tWTeG',
        (SELECT id FROM facilities WHERE code = 'RVDL'), 'super'),
    ('Elrond', 'Peredhel', 'EP', 'elrond@rivendell.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq', 
        (SELECT id FROM facilities WHERE code = 'RVDL'), 'admin'),
    ('Arwen', 'Evenstar', 'AE', 'arwen@rivendell.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'RVDL'), 'user'),

    -- Minas Tirith Staff
    ('Aragorn', 'Elessar', 'AE', 'aragorn@minas-tirith.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'MTGH'), 'super'),
    ('Faramir', 'Steward', 'FS', 'faramir@minas-tirith.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'MTGH'), 'user'),

    -- Lothlorien Staff
    ('Galadriel', 'Light', 'GL', 'galadriel@lothlorien.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'LOTH'), 'admin'),
    ('Celeborn', 'Elder', 'CE', 'celeborn@lothlorien.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'LOTH'), 'user'),

    -- Erebor Staff
    ('Thorin', 'Oakenshield', 'TO', 'thorin@erebor.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'ERBR'), 'super'),
    ('Balin', 'Fundin', 'BF', 'balin@erebor.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'ERBR'), 'user'),

    -- Shire Staff
    ('Bilbo', 'Baggins', 'BB', 'bilbo@shire.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'SHIR'), 'admin'),
    ('Frodo', 'Baggins', 'FB', 'frodo@shire.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'SHIR'), 'user')
ON CONFLICT (email) DO NOTHING;

-- Finally seed schedules
INSERT INTO schedules (rdos, anchor, controller_id) VALUES
    (ARRAY[6, 7], '2024-01-01', (SELECT id FROM controllers WHERE email = 'elrond@rivendell.med')),
    (ARRAY[5, 6], '2024-01-02', (SELECT id FROM controllers WHERE email = 'aragorn@minas-tirith.med')),
    (ARRAY[4, 5], '2024-01-03', (SELECT id FROM controllers WHERE email = 'galadriel@lothlorien.med')),
    (ARRAY[3, 4], '2024-01-04', (SELECT id FROM controllers WHERE email = 'thorin@erebor.med')),
    (ARRAY[2, 3], '2024-01-05', (SELECT id FROM controllers WHERE email = 'bilbo@shire.med')),
    (ARRAY[1, 2], '2024-01-06', (SELECT id FROM controllers WHERE email = 'frodo@shire.med')),
    (ARRAY[6, 7], '2024-01-07', (SELECT id FROM controllers WHERE email = 'arwen@rivendell.med')),
    (ARRAY[5, 6], '2024-01-08', (SELECT id FROM controllers WHERE email = 'faramir@minas-tirith.med')),
    (ARRAY[4, 5], '2024-01-09', (SELECT id FROM controllers WHERE email = 'celeborn@lothlorien.med')),
    (ARRAY[3, 4], '2024-01-10', (SELECT id FROM controllers WHERE email = 'balin@erebor.med'))
ON CONFLICT (controller_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Clean up in reverse order of dependencies
-- First remove schedules (due to foreign key to controllers)
DELETE FROM schedules
WHERE controller_id IN (
    SELECT id FROM controllers 
    WHERE email IN (
        'elrond@rivendell.med',
        'arwen@rivendell.med',
        'aragorn@minas-tirith.med',
        'faramir@minas-tirith.med',
        'galadriel@lothlorien.med',
        'celeborn@lothlorien.med',
        'thorin@erebor.med',
        'balin@erebor.med',
        'bilbo@shire.med',
        'frodo@shire.med'
    )
);

-- Then remove controllers (due to foreign key to facilities)
DELETE FROM controllers 
WHERE email IN (
    'elrond@rivendell.med',
    'arwen@rivendell.med',
    'aragorn@minas-tirith.med',
    'faramir@minas-tirith.med',
    'galadriel@lothlorien.med',
    'celeborn@lothlorien.med',
    'thorin@erebor.med',
    'balin@erebor.med',
    'bilbo@shire.med',
    'frodo@shire.med'
);

-- Finally remove facilities
DELETE FROM facilities 
WHERE code IN (
    'RVDL',
    'MTGH',
    'LOTH',
    'ERBR',
    'SHIR'
);
-- +goose StatementEnd
