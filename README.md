# kubectl-debugger-plugin

在 [TKE容器服务](https://cloud.tencent.com/document/product/457) 的集群中，该插件实现了下发 `debugger pod` 的能力，用于以下场景:

1、登录节点排查问题

2、登录节点查看其他非容器化进程的日志，比如日志组件


## 支持的节点类型

- 普通节点
- 原生节点
- 超级节点

## 支持的操作命令
```shell
# 适用超级节点
$ kubectl debugger pod <pod-name> -n <namespace> --rm

# 适用普通节点和原生节点
$ kubectl debugger node <node-name> --rm
```

## 超级节点场景特殊说明

超级节点上 `debugger pod` 并不是下发到了超级节点上，而是下发到了具体的pod上。

超级节点支持注解 `eks.tke.cloud.tencent.com/debug-pod: <pod-yaml>` , 用于下发其他 `pod` 到指定的 `pod` 上。

比如需要查看 `containerd` 系统组件的日志，参考如下操作：

```shell
$ chroot /host

$ journalctl -u containerd
```

## 要求的权限
```yaml
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "exec", "create", "delete"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get"]
```

## 效果展示
### pod
<img width="1143" alt="image" src="https://github.com/coderwangke/kubectl-debugger-plugin/assets/42019725/ece29b86-6e45-488b-bc5e-a948e154730c">

### node
<img width="865" alt="image" src="https://github.com/coderwangke/kubectl-debugger-plugin/assets/42019725/bfee8349-57f0-46e0-971c-565650fc0c16">

