# jkfr



# 简单示例

## JK-RPC-CLIENT

## JK-RPC-POOL

## JK-GRPC-CLIENT

## JK-GRPC-POOL

# 性能测试

## 不同核数机器(云主机)性能测试结果

### JK-RPC-CLIENT
![](images/JK-RPC-CLIENT.png)

### JK-RPC-POOL
![](images/JK-RPC-POOL.png)

### JK-GRPC-CLIENT
![](images/JK-GRPC-CLIENT.png)

### JK-GRPC-POOL
![](images/JK-GRPC-POOL.png)

## 连接池——不同连接数对请求响应性能的影响(8核机器)

由以下测试可看到，无论RPC还是gRPC，当连接池的连接数达到16时，就都差不多达到了请求响应性能的最大值，再增加更多的连接数对请求性能的提升就相对较小，甚至还会使得性能下降。因此，我们连接池的默认最大连接数就设置为了16.
### JK-RPC-POOL
![](images/JK-RPC-POOL-CONNS.png)
### JK-GRPC-POOL
![](images/JK-GRPC-POOL-CONNS.png)

## 不同负载均衡策略性能比较

  *1c-server，2c-server*是分别单独请求这两台机器时的性能测试

  - *1c-server* 指单核服务器，每秒能够响应**6W+**左右
  - *2c-server* 指的双核服务器，每秒能够响应**13W+**左右,可以看到基本**相当于1c-server的2倍**左右

  *round，random，least* 是指使用不同的负载均衡策略同时请求两台服务器的性能测试

  - *round* 轮询模式

  - *radom* 随机模式

  - *least* 最少请求模式

    从统计图表可以看到，*round，radom*模式的每秒请求响应数在12W左右，差不多相当于***1c-server* 的2倍**；而*least*模式的每秒请求响应数在20W左右，相当于***1c-server + 2c-server* 的每秒请求响应数**

    因此根据统计，整个集群的性能计算：

    *round，radom*模式： **总性能 = n * lowest-server**

    *least*模式: **总性能 = server1 + server2 + server3 ······**

    由此，我们可以看到，*round，radom*模式并不能很好的完全利用集群的性能，部分较好的机器可能会存在性能浪费，而性能不是那么好的机器可能又会存在过载的可能；而使用*least*模式就能很好的充分的利用整个集群的性能，能够智能将更多的请求发送到性能较好的机器或响应较快的机器。

![](images/BALANCER-STRATEGY.png)



