### 1、功能介绍
`in-cluster-tester`是利用kubernetes中的crd和operator实现的自动化测试框架。每个测试例作为一个pod运行在其中，并且最终会在相应的testcase实例中记录测试例的状态，同时会在summary实例中记录汇总状态。

该项目包含两个组件 —— 测试用例所在pod的`sidecar`容器和同步状态的`testcase-operator`
- sidecar是一个http server，提供以下3个接口：
   - 提供POST接口供测试用例容器写message
        - 测试用例容器调用该接口往sidecar中写入想打印的信息(无论错误信息还是正常信息)，最大1024字节
        - `curl -X POST   http://localhost:8099/message   -H 'cache-control: no-cache'   -H 'content-type: text/plain'   -d 1234`
   - 提供GET接口供operator获取测试用例容器写入sidecar的message
        - `curl -X GET http://localhost:8099/message`
   - 提供POST接口共operator给sidecar设置延迟退出周期，以使pod退出后，operator根据pod状态delete pod
        - `curl -X POST http://localhost:8099/delay?seconds=10`
- testcase-operator提供以下功能
   - watch主资源TestCase对象，为每个TestCase对象创建一个pod，运行起测试用例
   - watch次级资源Pod对象，当测试用例容器运行完获取exitcode，根据exitcode
        - 往TestCase对象中更新状态，并更新summary实例状态（汇总信息）
        - 往sidecar发送pod延迟退出周期，这个周期过后，pod会退出，operator watch pod状态为"Failed"或者"Succeeded"后，删除pod


### 2、deploy中文件与配置项介绍
- `rbac.yaml` —— `Cluster`角色`testcase-operator`的访问控制
- `deploy.sh` —— 初始部署脚本，部署`rbac`以及`testcase`和`summary`两个`crd`
- `operator.yaml` —— 主资源testcase、二级资源pod的控制器
    - `operator.yaml`中有如下配置
        - *SIDECAR_IMAGE* —— 测试用例pod中的sidecar容器的镜像
        - *SIDECAR_PORT* —— 测试用例pod中的sidecar容器http server监听的端口(目前为8099)
        - *DELAY_SECONDS_AFTER_TESETER_PASSED* —— 测试用例执行通过后多长时间删除测试用例pod
        - *DELAY_SECONDS_AFTER_TESETER_FAILED* —— 测试用例执行失败后多长时间删除测试用例po
        - *TESTCASE_SUMMARY* —— 要汇总所用测试例状态的summary实例
- `crds/testcase.yaml` —— `testcase`资源定义
    - spec
        - image —— 测试例镜像
        - commands —— 测试例容器要传入的命令
    - status
        - podName —— 该testcase对象对应的pod名（${NAMESPACE}/${NAME}）
        - result  —— 测试例运行结果（passed or failed）
        - message —— 测试例返回信息
        - completeTime —— 测试例执行完成时间
- `crds/testcase-example.yaml` —— `testcase`实例
- `crds/summary.yaml` —— `summary`资源定义
    - spec —— 无
    - status
        - totalNumber   ——  总测试例数量
        - passedNumber  ——  通过的测试例数量
        - failedNumber  ——  失败的测试例数量
- `crds/summary-instance.yaml` —— `summary`实例
    - name —— summary实例名，必须与`operator.yaml`中的`TESTCASE_SUMMARY`的值相同


### 3、使用说明
#### 3.1 部署
如果需要修改configmap和operator配置，请先在`operator-all-in-one.yaml`中进行修改，之后按以后步骤继续。
    ```bash
    tar zxvf deploy.tgz
    cd deploy
    kubectl create -f operator-all-in-one.yaml
    ```

#### 3.2 创建testcase实例
这里以`deploy/crds/testcase-example.yaml`为例，用户也可以自行构建，填入要执行的测试例的`image`和`commands`
```shell
kubectl create -f crds/testcase-example.yaml
```

#### 3.3 查看状态
- 可以通过以下命令查看多个测试例的汇总信息
    - testcase-summary为步骤1中的summary实例的名称
    ```bash
    kubectl get summ testcase-summary -o yaml
    ```
- 也可以通过以下命令查看单个测试例的状态
    ```bash
    kubectl get testcase example-testcase -o yaml
    ```
    

### 4 测试例编写
测试例可以参考`in-cluster-tester`项目的example目录中的代码