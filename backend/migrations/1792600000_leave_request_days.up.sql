-- Working days a leave request counts for, fixed when it is applied so that
-- approving or cancelling it later moves exactly the balance that was checked,
-- even if holidays change in between. Requests that exist before this
-- migration were counted in calendar days, so they keep that count.
ALTER TABLE leave_requests ADD COLUMN days INTEGER;
UPDATE leave_requests SET days = (to_date - from_date) + 1;
ALTER TABLE leave_requests ALTER COLUMN days SET NOT NULL;
ALTER TABLE leave_requests ADD CONSTRAINT leave_requests_days_positive CHECK (days > 0);
