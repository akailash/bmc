sudo docker-compose ps | grep Up | awk '{print $1}' | tr "\\n" " " | xargs --no-run-if-empty sudo docker stats
