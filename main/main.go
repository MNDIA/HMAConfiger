package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"
)


type HMAConfig struct {
    ConfigVersion  *int                    `json:"configVersion"`
    DetailLog      *bool                   `json:"detailLog"`
    MaxLogSize     *int                    `json:"maxLogSize"`
    ForceMountData *bool                   `json:"forceMountData"`
    Templates      map[string]Template     `json:"templates"`
    Scope          map[string]ScopeSetting `json:"scope"`
}


type Template struct {
    IsWhitelist *bool    `json:"isWhitelist"`
    AppList     []string `json:"appList"`
}


type ScopeSetting struct {
    UseWhitelist      *bool    `json:"useWhitelist"`
    ExcludeSystemApps *bool    `json:"excludeSystemApps"`
    ApplyTemplates    []string `json:"applyTemplates"`
    ExtraAppList      []string `json:"extraAppList"`
}

func main() {
    key := flag.String("key", "", "目录位置")
    command := flag.String("command", "", "执行命令: add 或 delete")
    packageName := flag.String("packageName", "", "应用包名")
    templateName := flag.String("templateName", "", "模板名称(add命令需要)")
    isWhitelistStr := flag.String("isWhitelist", "false", "是否为白名单(true/false)")
    excludeSystemAppsStr := flag.String("excludeSystemApps", "false", "是否排除系统应用(true/false)")
    flag.Parse()


    if *key == "" || *command == "" || *packageName == "" {
        fmt.Println("错误: 必须提供key, command和packageName参数")
        os.Exit(1)
    }

    path1 := "/data/data/com.tsng.hidemyapplist/files/config.json"
    path2 := fmt.Sprintf("/data/misc/hide_my_applist_%s/config.json", *key)

   

    switch *command {
    case "add":
        if *templateName == "" {
            fmt.Println("错误: add命令必须提供templateName参数")
            os.Exit(1)
        }
		isWhitelist := *isWhitelistStr == "true"
    	excludeSystemApps := *excludeSystemAppsStr == "true"
        if !isWhitelist && excludeSystemApps {
            fmt.Println("错误: 黑名单模式(isWhitelist=false)下excludeSystemApps必须为false")
            excludeSystemApps = false
        }
        addToConfig(path1, *packageName, *templateName, isWhitelist, excludeSystemApps)
        addToConfig(path2, *packageName, *templateName, isWhitelist, excludeSystemApps)
    case "delete":
        deleteFromConfig(path1, *packageName)
        deleteFromConfig(path2, *packageName)
    default:
        fmt.Printf("错误: 未知命令 '%s'. 请使用 'add' 或 'delete'\n", *command)
        os.Exit(1)
    }
}
func addToConfig(path string, packageName string, templateName string, isWhitelist bool, excludeSystemApps bool) {

    data, err := os.ReadFile(path)
    if err != nil {
        fmt.Printf("读取配置文件失败 %s: %v\n", path, err)
        return
    }

    var config HMAConfig
    err = json.Unmarshal(data, &config)
    if err != nil {
        fmt.Printf("解析配置文件失败 %s: %v\n", path, err)
        return
    }

    if config.Templates == nil {
        fmt.Printf("配置文件 %s 中没有templates部分\n", path)
        return
    }

    template, exists := config.Templates[templateName]
    if !exists {
        fmt.Printf("模板 '%s' 不存在于配置文件 %s 中\n", templateName, path)
        return
    }

    switch {
    case template.IsWhitelist == nil:
        fmt.Printf("模板 '%s' 在配置文件 %s 中未定义黑白名单类型\n", templateName, path)
        return
    case *template.IsWhitelist != isWhitelist:
        fmt.Printf("模板 '%s' 的黑白名单类型(%v)与请求不符(%v)\n", 
                  templateName, *template.IsWhitelist, isWhitelist)
        return
    }

    if config.Scope == nil {
        config.Scope = make(map[string]ScopeSetting)
    }

    scopeSetting := ScopeSetting{
        UseWhitelist:      &isWhitelist,
        ExcludeSystemApps: &excludeSystemApps,
        ApplyTemplates:    []string{templateName},
        ExtraAppList:      []string{},
    }

    config.Scope[packageName] = scopeSetting

    updatedData, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        fmt.Printf("序列化配置失败: %v\n", err)
        return
    }

    err = os.WriteFile(path, updatedData, 0644)
    if err != nil {
        fmt.Printf("写入配置文件失败 %s: %v\n", path, err)
        return
    }

    fmt.Printf("成功添加包名 '%s' 到配置 '%s'\n", packageName, path)
}

func deleteFromConfig(path string, packageName string) {
    data, err := os.ReadFile(path)
    if err != nil {
        fmt.Printf("读取配置文件失败 %s: %v\n", path, err)
        return
    }

    var config HMAConfig
    err = json.Unmarshal(data, &config)
    if err != nil {
        fmt.Printf("解析配置文件失败 %s: %v\n", path, err)
        return
    }

    switch {
    case config.Scope == nil:
        fmt.Printf("配置文件 %s 中没有scope部分\n", path)
        return
    case _, exists := config.Scope[packageName]; !exists:
        fmt.Printf("包名 '%s' 不在配置 %s 中\n", packageName, path)
        return
    }
    delete(config.Scope, packageName)

    updatedData, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        fmt.Printf("序列化配置失败: %v\n", err)
        return
    }

    err = os.WriteFile(path, updatedData, 0644)
    if err != nil {
        fmt.Printf("写入配置文件失败 %s: %v\n", path, err)
        return
    }

    fmt.Printf("成功从配置 '%s' 中删除包名 '%s'\n", path, packageName)
}