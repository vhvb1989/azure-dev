# Using azd environment from hooks

As an azd templates author, you might need to extend what your template can do by adding azd hooks. While hooks are very simple way to provide your own scripts and operations, there are a few details to keep in mind when your scripts will be interacting with azd's state, like the values from the environment (either just reading or setting values). In this article, you will learn some of the options you have for writing hooks, as well as interacting with the azd environment.

## Azd environment

In a few words, azd combines the `state` and the `configuration` for your application into one folder and calls it `environment`.

**The configuration** is a set of inputs to define how to create your application's resources or how services behave. For example, you can have a setting to define the `SKU` to be use for hosting one of your services, setting a `basic` SKU by default and expecting customers to set the right SKU by using that configuration.

**The state** are a set of values to describe properties for your application at a given point in time. For example, after calling `azd provision` all the infrastructure deployment's outputs are persisted in the environment. These values will not define or change the behavior of the application.

Azd allows you to create and switch between multiple environments. Each environment is created in a folder called `.azure`, next to the `azure.yaml` in your project. You can use `azd env --help` to discover how to create, list or switch between environments.

> !Note: Azd environment is different from CI/CD environments. **Do not** think about azd environment as GitHub environments.

When you run azd commands, azd uses the default selected environment. You can see the current default environment running `azd env list`. And you can override and set the environment to be use with:

 - **AZURE_ENV_NAME**, You can set the name of the environment in your system's variables and azd will use it instead of the default selected environment.
 - **-e**, You can use the `-e` flag when running azd commands to define the name of the environment to use. Azd will ignore your system's variables and the default selected environment when -e flag is set.

 > If you want to manually change the default selected environment without running `azd env select <name>`, you can update the `.azure/config` file directly.

## Azd environment in hooks

When azd invokes a hook, the environment values are automatically injected and you can read them the same way you would read any other value from your terminal. Here is a very simple example that demonstrates this. Consider the next hook definition:

```yaml
name: demo-hooks
hooks:
  preprovision:
    shell: sh
    run: ./script.sh
```

Now let's create `script.sh` like:

```bash
#!/bin/bash

echo $ENV_VAR_FROM_AZD

# just to let you see the output
sleep 5
```

And before you test it, make sure you set a value running:

```bash
azd env set ENV_VAR_FROM_AZD "hello azd"
# just to verify the value is in azd environment
azd env get-values
# you should see the `ENV_VAR_FROM_AZD="hello azd"` listed
```

Now you can run `azd provision` to trigger the hook. Or you can use `azd hooks run preprovision` so you will only execute the hook alone. You will see `hello azd` in the screen, as the `$ENV_VAR_FROM_AZD` is injected to the hooks execution. You can verify that if you manually run the script `./script.sh`, it won't display the message, as `ENV_VAR_FROM_AZD` won't have any value.

So, as you are creating a new hook and testing it, make sure to use `azd hooks run <hook name>` if you are reading azd environment. Otherwise, if you try manually running the script, you won't see the azd environment values.

In case you want to reproduce the same example in Windows powershell:

```yaml
name: demo-hooks
hooks:
  preprovision:
    shell: pwsh
    run: ./script.ps1
```

Now let's create `script.ps1` like:

```pwsh
Write-Output $env:ENV_VAR_FROM_AZD

# just to let you see the output
sleep 5
```

### Self environment injection

In case you want to be able to *manually run* the `script` from a hook and without using azd, but you still want to read azd environment.

You might want to allow customers to manually run the scripts in case the hook failed when it was invoked by azd. For example, as a troubleshooting guide, you might be telling folks which script to run to correct missing settings or repair application components. For such cases, consider asking customers to use `azd hooks run <hook name>` as a way to make it simpler for customers and still delegating the environment injection to azd.

But, if you still want to allow the manual invocation, you will need to teach your script how to pull pull the environment values from azd by calling and parsing the output of `azd env get-values`. Below are 2 strategies you can implement on your script.

#### Use script scoped variables

Ideally, the azd environment should be read and used by the script without affecting any other process or the terminal session. See the next code:

```bash
#!/bin/bash

declare -A azdEnv
while IFS='=' read -r key value; do
    value=$(echo "$value" | sed 's/^"//' | sed 's/"$//')
    azdEnv[$key]=$value
done <<EOF
$(azd env get-values)
EOF

echo "${azdEnv["ENV_VAR_FROM_AZD"]}"

# just to let you see the output
sleep 5
```

By using `declare -A azdEnv`, you can allocate a set to write the keys and values from the `azd env get-values` output. Then you can access the values by referencing the environment keys with `"${azdEnv["ENV_VAR_FROM_AZD"]}"`. This strategy would allow you to test or run your script either manually or using `azd hooks run <hook name>`, and without affecting the terminal variables. 

Here is the powershell equivalent.

```pwsh
$azdEnv = @{}
$azdEnvRaw = azd env get-values

foreach ($line in $azdEnvRaw -split "`n") {
    if ($line -match '^(.*?)=(.*)$') {
        $key = $matches[1]
        $value = $matches[2] -replace '^"|"$'
        $azdEnv[$key] = $value
    }
}

Write-Output $azdEnv["ENV_VAR_FROM_AZD"]

# just to let you see the output
Start-Sleep -Seconds 5
```

This strategy is fine and easy to implement when you have only one script, but if you have many scripts, you would need to copy and paste the same code in all the scripts, which might be hard to maintain. For such scenarios, take a look to the next strategy.

#### Use terminal scoped variables

When you have multiple scripts and you need all of some of them to access azd environment, you can make your script to set the azd environment in your terminal variables. You have to be careful with this strategy and ideally make sure to restore your terminal variables once all your scripts have run. This is because some terminals like powershell does not create one environment for executing the script, but update the environment of the terminal you are running in.

The main benefit of using your terminal variables is that you just need to set the variables one time, at your starting point, and then all scripts you run after will read the azd environment as direct variables from the terminal. See the next example:

```bash
#!/bin/bash

declare -A azdEnv
while IFS='=' read -r key value; do
    value=$(echo "$value" | sed 's/^"//' | sed 's/"$//')
    export "$key=$value"
done <<EOF
$(azd env get-values)
EOF

echo $ENV_VAR_FROM_AZD

# just to let you see the output
sleep 5
```

And the powershell version:

```pwsh
foreach ($line in (& azd env get-values)) {
    if ($line -match "([^=]+)=(.*)") {
        $key = $matches[1]
        $value = $matches[2] -replace '^"|"$'
	    [Environment]::SetEnvironmentVariable($key, $value)
    }
}

Write-Output $env:ENV_VAR_FROM_AZD

# just to let you see the output
Start-Sleep -Seconds 5
```

After running your script, make sure that azd environment is not leaked and persisted to your terminal variables, as it might affect future execution of azd commands because azd won't be switching to any other environment (unless you force it with the -e flag). This is because the `AZURE_ENV_NAME` would be persisted as terminal variable. So, depending on the terminal you are using (for sure on powershell), you would need to have a `restore variables` script at the end.