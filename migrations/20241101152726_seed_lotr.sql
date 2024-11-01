-- +goose Up
-- +goose StatementBegin
-- First, create a facility
INSERT INTO facilities (name, code) 
VALUES ('Minas Tirith ARTCC', 'MTIR') 
RETURNING id AS facility_id;

-- Create basic roles
INSERT INTO roles (name) 
VALUES 
    ('Administrator'),
    ('Controller')
RETURNING id, name;

-- Create two controllers
INSERT INTO controllers (name, initials, email, facility_id) 
VALUES 
    ('Gandalf Grey', 'GG', 'gandalf@white-tower.mt', 
        (SELECT id FROM facilities WHERE code = 'MTIR')),
    ('Pippin Took', 'PT', 'pippin@citadel.mt',
        (SELECT id FROM facilities WHERE code = 'MTIR'))
RETURNING id, name;

-- Assign roles to controllers at their facilities
INSERT INTO controller_facility_roles (controller_id, facility_id, role_id)
VALUES 
    (
        (SELECT id FROM controllers WHERE email = 'gandalf@white-tower.mt'),
        (SELECT id FROM facilities WHERE code = 'MTIR'),
        (SELECT id FROM roles WHERE name = 'Administrator')
    ),
    (
        (SELECT id FROM controllers WHERE email = 'pippin@citadel.mt'),
        (SELECT id FROM facilities WHERE code = 'MTIR'),
        (SELECT id FROM roles WHERE name = 'Controller')
    );

-- Optional: Create schedules for both controllers
INSERT INTO schedules (name, rdos, anchor, controller_id)
VALUES 
    (
        'White Council Schedule',
        ARRAY[0, 1],  -- Saturday and Sunday off
        '2024-01-01', -- Anchor date
        (SELECT id FROM controllers WHERE email = 'gandalf@white-tower.mt')
    ),
    (
        'Guard of the Citadel Schedule',
        ARRAY[2, 3],  -- Monday and Tuesday off
        '2024-01-01', -- Anchor date
        (SELECT id FROM controllers WHERE email = 'pippin@citadel.mt')
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Clean up in reverse order of dependencies
DELETE FROM schedules 
WHERE controller_id IN (
    SELECT id FROM controllers WHERE email IN ('gandalf@white-tower.mt', 'pippin@citadel.mt')
);

DELETE FROM controller_facility_roles 
WHERE controller_id IN (
    SELECT id FROM controllers WHERE email IN ('gandalf@white-tower.mt', 'pippin@citadel.mt')
);

DELETE FROM controllers 
WHERE email IN ('gandalf@white-tower.mt', 'pippin@citadel.mt');

DELETE FROM roles 
WHERE name IN ('Administrator', 'Controller');

DELETE FROM facilities 
WHERE code = 'MTIR';
-- +goose StatementEnd
