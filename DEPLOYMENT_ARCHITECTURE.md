# 游戏访问机制部署架构文档

## 🌐 整体部署架构

### 生产环境架构图

```mermaid
graph TB
    subgraph "用户访问层"
        A[用户] --> B[集成商系统]
        B --> C[负载均衡器<br/>LB]
    end
    
    subgraph "API网关层"
        C --> D[API网关<br/>Kong/Nginx]
        D --> E[限流控制]
        D --> F[身份验证]
        D --> G[路由分发]
    end
    
    subgraph "应用服务层"
        G --> H[认证服务集群<br/>3个节点]
        G --> I[游戏服务集群<br/>5个节点]
        G --> J[游戏分发服务集群<br/>2个节点]
    end
    
    subgraph "数据层"
        H --> K[Redis缓存集群<br/>3主3从]
        I --> K
        J --> K
        
        H --> L[MySQL主数据库]
        I --> L
        J --> L
        
        H --> M[MySQL从数据库集群<br/>2个节点]
        I --> M
        J --> M
        
        H --> N[MongoDB集群<br/>3节点]
        I --> N
        J --> N
    end
    
    subgraph "存储和CDN层"
        J --> O[CDN网络<br/>CloudFlare]
        O --> P[静态资源存储<br/>S3/OSS]
    end
    
    subgraph "监控和运维层"
        Q[监控系统<br/>Prometheus]
        R[日志系统<br/>ELK Stack]
        S[告警系统<br/>AlertManager]
        T[配置中心<br/>Consul]
    end
    
    H --> Q
    I --> Q
    J --> Q
    
    H --> R
    I --> R
    J --> R
    
    Q --> S
    R --> S
    
    H --> T
    I --> T
    J --> T
    
    style D fill:#e1f5ff
    style H fill:#ffe1f5
    style I fill:#e1ffe1
    style J fill:#fff4e1
    style K fill:#f5e1ff
    style L fill:#ffe1e1
```

## 💻 服务器资源配置

### 生产环境配置

| 服务 | 实例数量 | CPU | 内存 | 硬盘 | 操作系统 |
|------|----------|-----|------|------|----------|
| **负载均衡器** | 2 | 4核 | 8GB | 100GB SSD | Ubuntu 22.04 |
| **API网关** | 2 | 8核 | 16GB | 200GB SSD | Ubuntu 22.04 |
| **认证服务** | 3 | 8核 | 16GB | 200GB SSD | Ubuntu 22.04 |
| **游戏服务** | 5 | 16核 | 32GB | 500GB SSD | Ubuntu 22.04 |
| **游戏分发服务** | 2 | 8核 | 16GB | 200GB SSD | Ubuntu 22.04 |
| **Redis集群** | 6 | 8核 | 32GB | 200GB SSD | Ubuntu 22.04 |
| **MySQL主库** | 1 | 16核 | 64GB | 1TB SSD | Ubuntu 22.04 |
| **MySQL从库** | 2 | 8核 | 32GB | 1TB SSD | Ubuntu 22.04 |
| **MongoDB集群** | 3 | 8核 | 32GB | 500GB SSD | Ubuntu 22.04 |
| **监控服务器** | 1 | 8核 | 16GB | 200GB SSD | Ubuntu 22.04 |
| **日志服务器** | 1 | 8核 | 16GB | 500GB SSD | Ubuntu 22.04 |

## 🔄 负载均衡配置

### Nginx 负载均衡配置

```nginx
upstream auth_service {
    least_conn;
    server auth-1.internal:8080 weight=3;
    server auth-2.internal:8080 weight=3;
    server auth-3.internal:8080 weight=3;
    
    # 健康检查
    check interval=3000 rise=2 fall=3 timeout=1000;
    
    # 备份节点
    server auth-backup.internal:8080 backup;
}

upstream game_service {
    ip_hash;  # 会话保持
    server game-1.internal:8081 weight=2;
    server game-2.internal:8081 weight=2;
    server game-3.internal:8081 weight=2;
    server game-4.internal:8081 weight=2;
    server game-5.internal:8081 weight=2;
    
    # 健康检查
    check interval=3000 rise=2 fall=3 timeout=1000;
}

upstream game_distribution {
    least_conn;
    server dist-1.internal:8082 weight=2;
    server dist-2.internal:8082 weight=2;
    
    # 健康检查
    check interval=3000 rise=2 fall=3 timeout=1000;
}
```

### API网关配置

```yaml
# Kong API Gateway 配置示例
services:
  - name: auth-service
    url: http://auth-service:8080
    routes:
      - name: auth-route
        paths:
          - /v1/auth
        strip_path: true
        
  - name: game-service
    url: http://game-service:8081
    routes:
      - name: game-route
        paths:
          - /v1/game
        strip_path: true
        
  - name: game-distribution
    url: http://game-distribution:8082
    routes:
      - name: distribution-route
        paths:
          - /v1/distribution
        strip_path: true

plugins:
  - name: rate-limiting
    config:
      minute: 100
      hour: 1000
      policy: redis
      redis_host: redis.internal
      redis_port: 6379
      
  - name: jwt
    config:
      uri_param_names:
        - token
      claims_to_verify:
        - exp
      key_claim_name: kid
      
  - name: cors
    config:
      origins:
        - "https://*.platform.com"
      methods:
        - GET
        - POST
        - PUT
        - DELETE
      headers:
        - Accept
        - Accept-Version
        - Content-Length
        - Content-MD5
        - Content-Type
```

## 🗄️ 数据库部署

### MySQL 主从复制配置

```sql
-- 主数据库配置 (my.cnf)
[mysqld]
server-id = 1
log-bin = mysql-bin
binlog-format = ROW
max-binlog-size = 100M
expire-logs-days = 7

-- 从数据库配置 (my.cnf)
[mysqld]
server-id = 2
relay-log = mysql-relay-bin
read-only = 1
relay-log-purge = 1
```

### Redis 集群配置

```bash
# Redis Sentinel 配置
port 6379
bind 0.0.0.0
protected-mode yes
requirepass your_strong_password_here

# 持久化配置
save 900 1
save 300 10
save 60 10000

# 集群配置
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
cluster-require-full-coverage yes

# Sentinel 配置
port 26379
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel parallel-syncs mymaster 1
sentinel failover-timeout mymaster 60000
sentinel auth-pass mymaster your_strong_password_here
```

### MongoDB 集群配置

```yaml
# MongoDB 分片配置
sharding:
  clusterRole: shardsvr

replication:
  replSetName: shard1

net:
  bindIp: 0.0.0.0
  port: 27017

security:
  keyFile: /etc/mongodb-keyfile
  authorization: enabled

systemLog:
  destination: file
  path: /var/log/mongodb/mongod.log
  logAppend: true

storage:
  dbPath: /var/lib/mongodb
  journal:
    enabled: true
```

## 🚀 容器化部署

### Docker Compose 配置

```yaml
version: '3.8'

services:
  # 负载均衡器
  nginx:
    image: nginx:1.21-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - api-gateway
    networks:
      - frontend

  # API网关
  api-gateway:
    image: kong:2.8-alpine
    ports:
      - "8000:8000"
      - "8443:8443"
    environment:
      KONG_DATABASE: "off"
      KONG_PROXY_ACCESS_LOG: /dev/stdout
      KONG_ADMIN_ACCESS_LOG: /dev/stdout
      KONG_PROXY_ERROR_LOG: /dev/stderr
      KONG_ADMIN_ERROR_LOG: /dev/stderr
      KONG_ADMIN_LISTEN: "0.0.0.0:8001"
    networks:
      - frontend
      - backend

  # 认证服务
  auth-service:
    image: platform/auth-service:latest
    ports:
      - "8080:8080"
    environment:
      NODE_ENV: production
      DATABASE_URL: mysql://user:pass@mysql-primary:3306/auth_db
      REDIS_URL: redis://redis-master:6379
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - mysql-primary
      - redis-master
    deploy:
      replicas: 3
    networks:
      - backend

  # 游戏服务
  game-service:
    image: platform/game-service:latest
    ports:
      - "8081:8081"
    environment:
      NODE_ENV: production
      DATABASE_URL: mysql://user:pass@mysql-primary:3306/game_db
      REDIS_URL: redis://redis-master:6379
      MONGO_URL: mongodb://mongo-primary:27017/game_data
    depends_on:
      - mysql-primary
      - redis-master
      - mongo-primary
    deploy:
      replicas: 5
    networks:
      - backend

  # MySQL 主数据库
  mysql-primary:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: platform_db
      MYSQL_USER: platform_user
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    volumes:
      - mysql-primary-data:/var/lib/mysql
      - ./mysql-primary.cnf:/etc/mysql/conf.d/custom.cnf:ro
    networks:
      - backend

  # MySQL 从数据库
  mysql-replica:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
    volumes:
      - mysql-replica-data:/var/lib/mysql
      - ./mysql-replica.cnf:/etc/mysql/conf.d/custom.cnf:ro
    depends_on:
      - mysql-primary
    networks:
      - backend

  # Redis 主节点
  redis-master:
    image: redis:7-alpine
    command: redis-server /usr/local/etc/redis/redis.conf
    volumes:
      - redis-master-data:/data
      - ./redis-master.conf:/usr/local/etc/redis/redis.conf:ro
    networks:
      - backend

  # MongoDB 主节点
  mongo-primary:
    image: mongo:6
    command: mongod --replSet rs0 --bind_ip_all
    volumes:
      - mongo-primary-data:/data/db
      - ./mongo-keyfile:/etc/mongo-keyfile:ro
    networks:
      - backend

networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge

volumes:
  mysql-primary-data:
  mysql-replica-data:
  redis-master-data:
  mongo-primary-data:
```

## 📊 监控和日志系统

### Prometheus 监控配置

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'auth-service'
    static_configs:
      - targets: ['auth-service:8080']
    metrics_path: '/metrics'
    
  - job_name: 'game-service'
    static_configs:
      - targets: ['game-service:8081']
    metrics_path: '/metrics'
    
  - job_name: 'mysql'
    static_configs:
      - targets: ['mysql-exporter:9104']
      
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
      
  - job_name: 'mongodb'
    static_configs:
      - targets: ['mongodb-exporter:9216']

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']
```

### 告警规则配置

```yaml
# alert_rules.yml
groups:
  - name: application_alerts
    interval: 30s
    rules:
      # 服务可用性告警
      - alert: ServiceDown
        expr: up{job=~"auth-service|game-service"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"
          description: "Service {{ $labels.job }} has been down for more than 1 minute."
          
      # 高错误率告警
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} for {{ $labels.job }}"
          
      # 高内存使用告警
      - alert: HighMemoryUsage
        expr: (container_memory_usage_bytes / container_spec_memory_limit_bytes) > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage detected"
          description: "Container {{ $labels.name }} is using {{ $value | humanizePercentage }} of memory."
          
      # 数据库连接数告警
      - alert: DatabaseConnectionsHigh
        expr: mysql_global_status_threads_connected / mysql_global_variables_max_connections > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connections high"
          description: "Database connections are {{ $value | humanizePercentage }} of maximum."
```

### ELK 日志系统配置

```yaml
# logstash.conf
input {
  file {
    path => "/var/log/auth-service/*.log"
    type => "auth-service"
    start_position => "beginning"
  }
  
  file {
    path => "/var/log/game-service/*.log"
    type => "game-service"
    start_position => "beginning"
  }
}

filter {
  if [type] == "auth-service" or [type] == "game-service" {
    grok {
      match => {
        "message" => "%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{GREEDYDATA:log_message}"
      }
    }
    
    date {
      match => ["timestamp", "ISO8601"]
    }
    
    # 提取错误日志
    if [level] == "ERROR" {
      mutate {
        add_field => { "alert_level" => "error" }
      }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "platform-logs-%{+YYYY.MM.dd}"
  }
  
  # 错误日志输出到单独的索引
  if [level] == "ERROR" {
    elasticsearch {
      hosts => ["elasticsearch:9200"]
      index => "platform-errors-%{+YYYY.MM.dd}"
    }
  }
}
```

## 🔐 安全部署配置

### 防火墙配置

```bash
#!/bin/bash
# firewall-setup.sh

# 清除现有规则
iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X

# 设置默认策略
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# 允许本地回环
iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -o lo -j ACCEPT

# 允许已建立的连接
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# SSH 访问 (仅允许特定IP)
iptables -A INPUT -p tcp -s 203.0.113.0/24 --dport 22 -j ACCEPT

# HTTP/HTTPS
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# 内部服务端口
iptables -A INPUT -p tcp -s 10.0.0.0/8 --dport 8080:8082 -j ACCEPT
iptables -A INPUT -p tcp -s 10.0.0.0/8 --dport 6379 -j ACCEPT
iptables -A INPUT -p tcp -s 10.0.0.0/8 --dport 3306 -j ACCEPT
iptables -A INPUT -p tcp -s 10.0.0.0/8 --dport 27017 -j ACCEPT

# 监控端口
iptables -A INPUT -p tcp -s 10.0.0.0/8 --dport 9090:9093 -j ACCEPT

# 日志记录
iptables -A INPUT -j LOG --log-prefix "IPTables-Dropped: " --log-level 4

# 保存规则
iptables-save > /etc/iptables/rules.v4
```

### SSL/TLS 配置

```bash
#!/bin/bash
# ssl-setup.sh

# 生成自签名证书 (测试环境)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/nginx/ssl/server.key \
  -out /etc/nginx/ssl/server.crt \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=Platform/OU=IT/CN=*.platform.com"

# 生成 Let's Encrypt 证书 (生产环境)
certbot certonly --nginx \
  -d api.platform.com \
  -d games.platform.com \
  --email security@platform.com \
  --agree-tos \
  --no-eff-email

# 配置自动续期
echo "0 2 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'" | crontab -
```

## 🚦 部署流程

### CI/CD 流程图

```mermaid
graph LR
    A[代码提交] --> B[代码检查<br/>Lint/Test]
    B --> C[构建镜像<br/>Docker Build]
    C --> D[推送镜像<br/>Registry Push]
    D --> E[部署到测试环境<br/>Test Deploy]
    E --> F{测试通过?}
    F -->|否| A
    F -->|是| G[部署到预发布<br/>Staging Deploy]
    G --> H{预发布测试?}
    H -->|否| A
    H -->|是| I[蓝绿部署<br/>Blue-Green Deploy]
    I --> J[健康检查<br/>Health Check]
    J --> K{健康检查通过?}
    K -->|否| I
    K -->|是| L[切换流量<br/>Traffic Switch]
    L --> M[监控观察<br/>Monitor]
    M --> N{部署成功?}
    N -->|否| O[回滚操作<br/>Rollback]
    N -->|是| P[部署完成<br/>Complete]
    O --> I
```

### 部署脚本

```bash
#!/bin/bash
# deploy.sh

set -e

# 配置变量
IMAGE_TAG=$1
ENVIRONMENT=$2
SERVICE_NAME=$3

if [ -z "$IMAGE_TAG" ] || [ -z "$ENVIRONMENT" ] || [ -z "$SERVICE_NAME" ]; then
    echo "Usage: ./deploy.sh <image_tag> <environment> <service_name>"
    exit 1
fi

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}开始部署服务: $SERVICE_NAME${NC}"
echo -e "${GREEN}环境: $ENVIRONMENT${NC}"
echo -e "${GREEN}镜像版本: $IMAGE_TAG${NC}"

# 1. 拉取最新镜像
echo -e "${YELLOW}拉取最新镜像...${NC}"
docker pull registry.platform.com/$SERVICE_NAME:$IMAGE_TAG

# 2. 停止旧容器
echo -e "${YELLOW}停止旧容器...${NC}"
docker-compose -f docker-compose.$ENVIRONMENT.yml stop $SERVICE_NAME

# 3. 备份旧版本
echo -e "${YELLOW}备份旧版本...${NC}"
BACKUP_DIR="/backup/$SERVICE_NAME/$(date +%Y%m%d_%H%M%S)"
mkdir -p $BACKUP_DIR
cp -r /opt/$SERVICE_NAME/config $BACKUP_DIR/

# 4. 启动新容器
echo -e "${YELLOW}启动新容器...${NC}"
IMAGE_TAG=$IMAGE_TAG docker-compose -f docker-compose.$ENVIRONMENT.yml up -d $SERVICE_NAME

# 5. 健康检查
echo -e "${YELLOW}执行健康检查...${NC}"
for i in {1..30}; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}健康检查通过！${NC}"
        break
    fi
    
    if [ $i -eq 30 ]; then
        echo -e "${RED}健康检查失败！开始回滚...${NC}"
        docker-compose -f docker-compose.$ENVIRONMENT.yml down
        cp -r $BACKUP_DIR/config /opt/$SERVICE_NAME/
        docker-compose -f docker-compose.$ENVIRONMENT.yml up -d
        exit 1
    fi
    
    echo "等待服务启动... ($i/30)"
    sleep 2
done

# 6. 清理旧镜像
echo -e "${YELLOW}清理旧镜像...${NC}"
docker image prune -f

echo -e "${GREEN}部署成功完成！${NC}"
```

## 🧪 灾难恢复方案

### 数据备份策略

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup"
DATE=$(date +%Y%m%d)
RETENTION_DAYS=7

# MySQL 备份
echo "开始 MySQL 备份..."
mysqldump -u root -p$MYSQL_PASSWORD --all-databases | gzip > $BACKUP_DIR/mysql_$DATE.sql.gz

# Redis 备份
echo "开始 Redis 备份..."
redis-cli --rdb $BACKUP_DIR/redis_$DATE.rdb

# MongoDB 备份
echo "开始 MongoDB 备份..."
mongodump --host localhost --port 27017 --out $BACKUP_DIR/mongodb_$DATE

# 应用配置备份
echo "开始配置备份..."
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /opt/*/config/

# 上传到云存储
aws s3 sync $BACKUP_DIR s3://platform-backup/$DATE/

# 清理旧备份
find $BACKUP_DIR -name "*.gz" -mtime +$RETENTION_DAYS -delete
find $BACKUP_DIR -name "*.rdb" -mtime +$RETENTION_DAYS -delete
```

### 恢复流程

```bash
#!/bin/bash
# restore.sh

BACKUP_DATE=$1
BACKUP_DIR="/backup"

if [ -z "$BACKUP_DATE" ]; then
    echo "Usage: ./restore.sh <backup_date>"
    exit 1
fi

# 从云存储下载备份
aws s3 sync s3://platform-backup/$BACKUP_DATE/ $BACKUP_DIR/

# 停止服务
docker-compose -f docker-compose.production.yml down

# 恢复 MySQL
echo "恢复 MySQL 数据..."
gunzip < $BACKUP_DIR/mysql_$BACKUP_DATE.sql.gz | mysql -u root -p$MYSQL_PASSWORD

# 恢复 Redis
echo "恢复 Redis 数据..."
cp $BACKUP_DIR/redis_$BACKUP_DATE.rdb /var/lib/redis/dump.rdb
systemctl restart redis

# 恢复 MongoDB
echo "恢复 MongoDB 数据..."
mongorestore --host localhost --port 27017 --drop $BACKUP_DIR/mongodb_$BACKUP_DATE

# 恢复配置
echo "恢复配置文件..."
tar -xzf $BACKUP_DIR/config_$BACKUP_DATE.tar.gz -C /

# 启动服务
docker-compose -f docker-compose.production.yml up -d

echo "恢复完成！"
```

## 📝 运维清单

### 日常运维检查项

- [ ] 服务器资源使用情况检查
- [ ] 数据库备份执行情况
- [ ] 磁盘空间使用检查
- [ ] 应用日志异常检查
- [ ] 网络连接状态检查
- [ ] SSL证书有效期检查
- [ ] 监控告警规则验证

### 周期性维护任务

| 任务类型 | 频率 | 负责人 |
|----------|------|--------|
| 安全补丁更新 | 每周 | 安全团队 |
| 数据库性能优化 | 每月 | DBA |
| 系统日志分析 | 每月 | 运维团队 |
| 灾备演练 | 季度 | 运维团队 |
| 容量规划评估 | 季度 | 架构师 |

---

**文档版本**: v1.0  
**创建时间**: 2024-01-15  
**最后更新**: 2024-01-15  
**维护团队**: 平台运维团队