# K9s Dashboard

Jackal vendors in [K9s](https://k9scli.io/), a terminal-based UI to interact with your Kubernetes cluster. K9s is not necessary to deploy, manage, or operate Jackal or its deployed packages, but it is a great tool to use when you want to interact with your cluster. Since Jackal vendors in this tool, you don't have to worry about additional dependencies or trying to install it yourself!

## Using the k9s Dashboard

All you need to use the k9s dashboard is to:

1. Have access to a running cluster kubecontext
1. Have a jackal binary installed

Using the k9s Dashboard is as simple as using a single command!

```bash
jackal tools k9s
```

**Example k9s Dashboard**
![k9s dashboard](../.images/dashboard/k9s_dashboard_example.png)

More instructions on how to use k9s can be found on their [documentation site](https://k9scli.io/topics/commands/).
