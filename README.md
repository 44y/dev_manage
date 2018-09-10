## 目录说明

```bash
.
├── common.mk                           ### makefile 公共部分
├── deploy                              ### 部署文件目录
│   ├── dev.sh                          ##! 部署dev版本的部署脚本
│   ├── docker-compose-dev.yml          ##! dev版本对应的docker-compose.yml
│   ├── docker-compose-release.yml      ##! release版本对应的docker-compose.yml
│   ├── docker-compose-test.yml         ##! test版本对应的docker-compose.yml
│   ├── release.sh                      ##! 部署release版本的部署脚本
│   └── test.sh                         ##! 部署test版本的部署脚本
├── doc                                 ### 文档目录目录
│   └── design.md
├── docker                              ### 存放docker构建文件的目录
│   └── devices                         ##! devices docker
│       ├── devices.sh
│       ├── Dockerfile                  #!! 构建devices docker的Dockerfile
│       └── Makefile                    #!! 内层Makefile，编译docker，部署docker用
├── Makefile                            ### 最外层Makefile，需实现build,docker，deploy等目标
├── README.md                           ### 本文档
├── src                                 ### 存放需要编译的源代码目录
│   └── devices                         ##! devices主程序代码目录
│       ├── devices.go
│       └── Makefile                    #!! 内层Makefile，编译源代码用
├── test                                ### 测试用例目录
│   └── test.sh                         ##! 测试用例脚本
└── toolchain                           ### 工具链目录
    └── Dockerfile                      ##! 生成工具链的容器的Dockerfile

```


## 最外层Makefile

***make 目标说明***

|        目标            |必要性 |          参数              |            说明                              |
|-----------------------|------|---------------------------|---------------------------------------------|
| all/build (缺省为空)   | 必须  |         无                |编译代码，生成目标机程序；可以空实现                 |
| clean                 | 必须  |         无                |清理编译过程生成的文件；可以空实现                  |
| docker                | 必须  |DOCKER_PATH=? (缺省为dev)   |依赖all/build,构建docker 镜像                   |
| deploy                | 必须  |DOCKER_PATH=? (缺省为dev) DEPLOY_TARGET=? (缺省为dev)  |上传镜像；部署docker-compose.yml   |
| tools                 | 必须  |         无                |生成编译工具等；可以空实现                         |
| save                  | 可选  |DOCKER_PATH=? (缺省为dev)   |把本地的docker image压缩打包                     |
| clean-none            | 可选  |         无                |清理本地所有<none>镜像                           |
| run                   | 可选  |DOCKER_PATH=? (缺省为dev)   |本地启动指定镜像进行测试                           |
| test                  | 可选  |         无                |启动测试环境或进行单元测试                          |

***主要变量说明***

|        目标            |必要性  |        说明                                               |
|-----------------------|-------|----------------------------------------------------------|
| SUBDIRS               | 可选   |   源代码的子目录列表，如：src/hello                           |
| DOCKER_SUBDIRS        | 必须   |   构建docker的子目录列表,如：docker/hello docker/bye         |
| PKG_VERSION           | 可选   |   包（镜像）版号，如果有多个docker镜像时，不要使用该变量          |
| PKG_NAME              | 可选   |   包（镜像）名，如果有多个docker镜像时，不要使用该变量            |
| OUT_PUT               | 可选   |   导出给内存Makefile使用的bin输出路径                         |
| BUILD_VERSION         | 可选   |   导出给内存Makefile使用的版本号                              |
| DEPLOY_TARGET         | 必须   |   部署的目标，即指定deploy目录下的部署脚本名（不含.sh）           |


## 公共common.mk

***主要变量说明***

|        目标            |必要性 |        说明                                              |
|-----------------------|------|---------------------------------------------------------|
| DOCKER_REPERTORY      | 必须  |  上传docker的私有仓库，可包含端口                            |
| DOCKER_PATH           | 必须  |  docker的存放路径，可由外部传入,例如　make DOCKER_PATH=test   |



## 内层src目录Makefile

***make 目标说明***

|        目标            |必要性 |          参数              |            说明                          |
|-----------------------|------|---------------------------|------------------------------------------|
| all/build (缺省为空)   | 必须  |         无                 |  编译代码，生成目标机程序                    |
| clean                 | 必须  |         无                |  清理编译过程生成的文件                      |
| run                   | 可选  |DOCKER_PATH=? (缺省为dev)   |  本地启动指定镜像进行测试                    |
| test                  | 可选  |         无                |  启动测试环境                              |

***主要变量说明***

|        目标            |必要性 |        说明                                    |
|-----------------------|------|------------------------------------------------|
| BUILD_VERSION         | 必须  |  服务的版本号，被最外层的版本号覆盖                  |
| BUILD_TIME            | 必须  |  服务的编译时间                                  |
| BUILD_NAME            | 必须  |  服务名字                                       |
| COMMIT_SHA1           | 可选  |  提交的git commit id                            |
| SOURCE                | 可选  |  要编译的原代码列表                               |


## 内层docker目录Makefile

***make 目标说明***

|        目标            |必要性 |          参数              |            说明                             |
|-----------------------|------|---------------------------|---------------------------------------------|
| all/docker (缺省为空)  | 必须  |DOCKER_PATH=? (缺省为dev)   |构建docker 镜像                                |
| deploy                | 必须  |DOCKER_PATH=? (缺省为dev)   |依赖docker,上传镜像；部署docker-compose.yml待定  |
| save                  | 可选  |DOCKER_PATH=? (缺省为dev)   |把本地的docker image压缩打包                    |
| run                   | 可选  |DOCKER_PATH=? (缺省为dev)   |本地启动指定镜像进行测试                          |
| test                  | 可选  |         无                |启动测试环境或进行单元测试                        |



***主要变量说明***

|        目标            |必要性 |        说明                   |
|-----------------------|------|-------------------------------|
| DOCKER_IMG_NAME       | 必须  |  docker镜像名                  |
