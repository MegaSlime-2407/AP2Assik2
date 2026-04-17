CREATE OR REPLACE FUNCTION notify_order_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        PERFORM pg_notify('order_status_changed',
            json_build_object('order_id', NEW.id, 'status', NEW.status)::text
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'order_status_trigger'
    ) THEN
        CREATE TRIGGER order_status_trigger
        AFTER UPDATE ON orders
        FOR EACH ROW
        EXECUTE FUNCTION notify_order_status_change();
    END IF;
END $$;
