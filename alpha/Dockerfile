FROM mysql:5.7.30

# Update and clean apt cache
RUN apt-get clean && apt-get update --fix-missing && \
    # Install utils \
    apt install -y nano ssh inetutils-ping net-tools htop telnet curl redis-tools ncdu rsync whois dnsutils git zip gosu wget percona-toolkit vim;

# MySQL Custom Configs
COPY my.cnf /etc/mysql/my.cnf
COPY mysql.conf.d/ /etc/mysql/mysql.conf.d/
COPY my-init.sql /docker-entrypoint-initdb.d/
COPY users_data.sql /home/

# Touching log files
RUN touch -a /var/log/mysql/error.log && chown mysql:adm /var/log/mysql/error.log
RUN touch -a /var/log/mysql/general.log && chown mysql:adm /var/log/mysql/general.log
RUN touch -a /var/log/mysql/slow.log && chown mysql:adm /var/log/mysql/slow.log
