/***************************************************************
 * Validation functions
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

-- Validates that the given string is a valid FQDN (fully qualified domain name)
CREATE FUNCTION check_domain_fqdn(fqdn VARCHAR)
RETURNS BOOLEAN AS $$
BEGIN
    -- Quick length and null checks first
    IF fqdn IS NULL OR length(fqdn) < 3 OR length(fqdn) > 253 THEN
        RETURN false;
    END IF;

    -- Quick character set check (most restrictive first)
    IF fqdn !~ '^[a-zA-Z0-9.-]+$' THEN
        RETURN false;
    END IF;

    -- More expensive regex last
    IF fqdn !~* '^([a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$' THEN
        RETURN false;
    END IF;

    RETURN true;
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;

-- Validates that the given string is a valid mail address name (local part before the @)
CREATE FUNCTION check_mail_address_name(name VARCHAR)
RETURNS BOOLEAN AS $$
BEGIN
    -- Check length (max 64 chars for local part)
    IF name IS NULL OR length(name) < 1 OR length(name) > 64 THEN
        RETURN false;
    END IF;

    -- Disallow leading or trailing dot
    IF left(name, 1) = '.' OR right(name, 1) = '.' THEN
        RETURN false;
    END IF;

    -- Disallow consecutive dots
    IF name LIKE '%..%' THEN
        RETURN false;
    END IF;

    -- Allow only permitted characters: a-z, A-Z, 0-9, and these special chars: !#$%&'*+-/=?^_`{|}~. and dot
    IF name !~* '^[a-z0-9!#$%&''*+\-/=?^_`{|}~.]+$' THEN
        RETURN false;
    END IF;

    RETURN true;
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;
