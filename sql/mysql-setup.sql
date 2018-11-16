CREATE DATABASE sadmin;
CREATE USER 'sadmin_user'@'%' IDENTIFIED BY 'sadminpass' REQUIRE SSL;
GRANT ALL PRIVILEGES ON sadmin.* TO 'sadmin_user'@'%';
