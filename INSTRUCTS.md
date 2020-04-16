go run php.go --path="/path/to.php.script.php"

If you are not replicating, you can disable binlogging by changing your my.ini or my.cnf file. Open your my.ini or /etc/my.cnf (/etc/mysql/my.cnf), enter:

# vi /etc/my.cnf

Find a line that reads "log_bin" and remove or comment it as follows:

#log_bin = /var/log/mysql/mysql-bin.log

You also need to remove or comment following lines:

#expire_logs_days        = 10

#max_binlog_size         = 100M

Close and save the file. Finally, restart mysql server:

# service mysql restart