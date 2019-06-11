# Examples

## Make a report of images to be cleaned

You can generate a report of actions it would take, without actually doing any deletions.

```
$ ./bin/docker-registry-pruner -mode report -config ./config/gabe.yaml
{"level":"info","ts":1557265749.7190368,"caller":"cmd/main.go:51","msg":"Created Registry client for https://registry.company.net"}
{"level":"info","ts":1557265749.719176,"caller":"cmd/main.go:66","msg":"Loaded rule: Repos:tumblr/bb8,trash-island/docker-registry-pruner Selector{ignore tags [latest], match tags [^v\\d+\\.\\d+]} Action{keep latest 4 versions}"}
{"level":"info","ts":1557265749.719208,"caller":"cmd/main.go:71","msg":"Building image report for images: tumblr/bb8, trash-island/docker-registry-pruner"}
{"level":"info","ts":1557265749.719243,"caller":"cmd/main.go:123","msg":"Querying for manifests. This may take a while..."}
action image                               tag                       parsed_version           age_days
keep   tumblr/bb8                          v0.6.0-560-gb975605       0.6.0-560-gb975605       39
keep   tumblr/bb8                          v0.6.0-561-g63760b8       0.6.0-561-g63760b8       39
keep   tumblr/bb8                          v0.6.0-564-gbbda8c3       0.6.0-564-gbbda8c3       33
keep   tumblr/bb8                          v0.6.0-566-g82b147c       0.6.0-566-g82b147c       32
keep   trash-island/docker-registry-pruner v0.1.0-16-g710a0e3-master 0.1.0-16-g710a0e3-master 1
keep   trash-island/docker-registry-pruner v0.1.0-17-ge5bf735-master 0.1.0-17-ge5bf735-master 1
keep   trash-island/docker-registry-pruner v0.1.0-19-gdddaf1e-master 0.1.0-19-gdddaf1e-master 0
keep   trash-island/docker-registry-pruner v0.1.0-20-gf53cfa7-master 0.1.0-20-gf53cfa7-master 0
delete tumblr/bb8                          v0.6.0-497-g5820922       0.6.0-497-g5820922       123
delete tumblr/bb8                          v0.6.0-503-g8e24b6a       0.6.0-503-g8e24b6a       123
delete tumblr/bb8                          v0.6.0-511-g7127911       0.6.0-511-g7127911       123
delete tumblr/bb8                          v0.6.0-514-g39b0708       0.6.0-514-g39b0708       120
delete tumblr/bb8                          v0.6.0-516-gafffdf5       0.6.0-516-gafffdf5       119
delete tumblr/bb8                          v0.6.0-518-ga02e868       0.6.0-518-ga02e868       102
delete tumblr/bb8                          v0.6.0-521-g49e37cd       0.6.0-521-g49e37cd       98
delete tumblr/bb8                          v0.6.0-529-g7029e4c       0.6.0-529-g7029e4c       68
delete tumblr/bb8                          v0.6.0-531-g662a23d       0.6.0-531-g662a23d       47
delete tumblr/bb8                          v0.6.0-535-ge62b08a       0.6.0-535-ge62b08a       47
delete tumblr/bb8                          v0.6.0-547-ge2c98ab       0.6.0-547-ge2c98ab       40
delete trash-island/docker-registry-pruner v0.1.0-11-g0f017e0-master 0.1.0-11-g0f017e0-master 1
delete trash-island/docker-registry-pruner v0.1.0-14-g0a78f97-master 0.1.0-14-g0a78f97-master 1
delete trash-island/docker-registry-pruner v0.1.0-15-gad48256-master 0.1.0-15-gad48256-master 1
deleting 14 images, keeping 8 images
```

## Delete some shit

NOTE: make sure you are using the right config!!!!! This action will mutate your registry and potentially delete important things. Use `-mode report` first.

```
# lets delete some shit
$ ./bin/docker-registry-pruner -mode prune -config ./config/gabe.yaml
{"level":"info","ts":1557265970.763993,"caller":"cmd/main.go:51","msg":"Created Registry client for https://registry.company.net"}
{"level":"info","ts":1557265970.764477,"caller":"cmd/main.go:66","msg":"Loaded rule: Repos:tumblr/bb8,trash-island/docker-registry-pruner Selector{ignore tags [latest], match tags [^v\\d+\\.\\d+]} Action{keep latest 4 versions}"}
{"level":"info","ts":1557265970.76457,"caller":"cmd/main.go:74","msg":"Pruning tags for images: tumblr/bb8, trash-island/docker-registry-pruner"}
{"level":"info","ts":1557265970.7648351,"caller":"cmd/main.go:131","msg":"Querying for manifests. This may take a while..."}
{"level":"info","ts":1557265973.501079,"caller":"cmd/main.go:133","msg":"Beginning deletion of 14 images"}
{"level":"info","ts":1557265973.501157,"caller":"registry/client.go:98","msg":"9: deleting manifest for tumblr/bb8:v0.6.0-497-g5820922"}
{"level":"info","ts":1557265973.501204,"caller":"registry/client.go:98","msg":"3: deleting manifest for tumblr/bb8:v0.6.0-503-g8e24b6a"}
{"level":"info","ts":1557265973.501285,"caller":"registry/client.go:98","msg":"2: deleting manifest for tumblr/bb8:v0.6.0-516-gafffdf5"}
{"level":"info","ts":1557265973.501207,"caller":"registry/client.go:98","msg":"0: deleting manifest for tumblr/bb8:v0.6.0-511-g7127911"}
{"level":"info","ts":1557265973.5012422,"caller":"registry/client.go:98","msg":"8: deleting manifest for tumblr/bb8:v0.6.0-535-ge62b08a"}
{"level":"info","ts":1557265973.501249,"caller":"registry/client.go:98","msg":"6: deleting manifest for tumblr/bb8:v0.6.0-518-ga02e868"}
{"level":"info","ts":1557265973.5012622,"caller":"registry/client.go:98","msg":"4: deleting manifest for tumblr/bb8:v0.6.0-521-g49e37cd"}
{"level":"info","ts":1557265973.501269,"caller":"registry/client.go:98","msg":"1: deleting manifest for tumblr/bb8:v0.6.0-514-g39b0708"}
{"level":"info","ts":1557265973.5012732,"caller":"registry/client.go:98","msg":"5: deleting manifest for tumblr/bb8:v0.6.0-529-g7029e4c"}
{"level":"info","ts":1557265973.5012858,"caller":"registry/client.go:98","msg":"7: deleting manifest for tumblr/bb8:v0.6.0-531-g662a23d"}
{"level":"info","ts":1557265973.690368,"caller":"registry/client.go:104","msg":"8: manifest tumblr/bb8:v0.6.0-535-ge62b08a successfully deleted"}
{"level":"info","ts":1557265973.69044,"caller":"registry/client.go:98","msg":"8: deleting manifest for tumblr/bb8:v0.6.0-547-ge2c98ab"}
{"level":"info","ts":1557265973.6948102,"caller":"registry/client.go:98","msg":"5: deleting manifest for trash-island/docker-registry-pruner:v0.1.0-11-g0f017e0-master"}
{"level":"info","ts":1557265973.69545,"caller":"registry/client.go:98","msg":"7: deleting manifest for trash-island/docker-registry-pruner:v0.1.0-14-g0a78f97-master"}
{"level":"info","ts":1557265973.696696,"caller":"registry/client.go:98","msg":"2: deleting manifest for trash-island/docker-registry-pruner:v0.1.0-15-gad48256-master"}
{"level":"info","ts":1557265973.706157,"caller":"registry/client.go:104","msg":"6: manifest tumblr/bb8:v0.6.0-518-ga02e868 successfully deleted"}
{"level":"info","ts":1557265973.722357,"caller":"registry/client.go:104","msg":"0: manifest tumblr/bb8:v0.6.0-511-g7127911 successfully deleted"}
{"level":"info","ts":1557265973.7224078,"caller":"registry/client.go:104","msg":"4: manifest tumblr/bb8:v0.6.0-521-g49e37cd successfully deleted"}
{"level":"info","ts":1557265973.722363,"caller":"registry/client.go:104","msg":"3: manifest tumblr/bb8:v0.6.0-503-g8e24b6a successfully deleted"}
{"level":"info","ts":1557265973.803635,"caller":"registry/client.go:104","msg":"9: manifest tumblr/bb8:v0.6.0-497-g5820922 successfully deleted"}
{"level":"info","ts":1557265973.803627,"caller":"registry/client.go:104","msg":"1: manifest tumblr/bb8:v0.6.0-514-g39b0708 successfully deleted"}
{"level":"info","ts":1557265973.888031,"caller":"registry/client.go:104","msg":"8: manifest tumblr/bb8:v0.6.0-547-ge2c98ab successfully deleted"}
{"level":"info","ts":1557265974.265267,"caller":"registry/client.go:104","msg":"2: manifest trash-island/docker-registry-pruner:v0.1.0-15-gad48256-master successfully deleted"}
{"level":"info","ts":1557265974.4662452,"caller":"cmd/main.go:135","msg":"Deleted 9 images, encountered 5 errors"}
....

# hey, its been deleted!

[17:52:54] [gabec@Gabes-MacBook-Pro] ( 2 ) ~/code/tumblr/docker-registry-pruner master*
-> $ ./bin/docker-registry-pruner -mode report -config ./config/gabe.yaml
{"level":"info","ts":1557265982.717484,"caller":"cmd/main.go:51","msg":"Created Registry client for https://registry.company.net"}
{"level":"info","ts":1557265982.717578,"caller":"cmd/main.go:66","msg":"Loaded rule: Repos:tumblr/bb8,trash-island/docker-registry-pruner Selector{ignore tags [latest], match tags [^v\\d+\\.\\d+]} Action{keep latest 4 versions}"}
{"level":"info","ts":1557265982.717599,"caller":"cmd/main.go:71","msg":"Building image report for images: tumblr/bb8, trash-island/docker-registry-pruner"}
{"level":"info","ts":1557265982.7176151,"caller":"cmd/main.go:123","msg":"Querying for manifests. This may take a while..."}
action image                               tag                       parsed_version           age_days
keep   tumblr/bb8                          v0.6.0-560-gb975605       0.6.0-560-gb975605       39
keep   tumblr/bb8                          v0.6.0-561-g63760b8       0.6.0-561-g63760b8       39
keep   tumblr/bb8                          v0.6.0-564-gbbda8c3       0.6.0-564-gbbda8c3       33
keep   tumblr/bb8                          v0.6.0-566-g82b147c       0.6.0-566-g82b147c       32
keep   trash-island/docker-registry-pruner v0.1.0-16-g710a0e3-master 0.1.0-16-g710a0e3-master 1
keep   trash-island/docker-registry-pruner v0.1.0-17-ge5bf735-master 0.1.0-17-ge5bf735-master 1
keep   trash-island/docker-registry-pruner v0.1.0-19-gdddaf1e-master 0.1.0-19-gdddaf1e-master 1
keep   trash-island/docker-registry-pruner v0.1.0-20-gf53cfa7-master 0.1.0-20-gf53cfa7-master 0
deleting 0 images, keeping 8 images
```

## More config examples

See our configuration examples at [config/examples/](/config/examples/). These illustrate different configurations that can perform a number of useful retention actions.
