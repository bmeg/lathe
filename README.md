
# Lathe
Dynamic build system. 

Lathe allows a user to programatically define a workflow of tasks. 

The lathe build files are written in JavaScript. 

Example:
```javascript

prep = lathe.Workflow("prep")

projects = [
    {name: "BeatAML_2018"},
    {name: "FIMM_2016"},
    {name: "NCI60_2021"},
    {name: "Tavor_2020"},
    {name: "CCLE_2015"},
    {name: "GBM_scr2"},
    {name: "PDTX_2019"},
    {name: "UHNBreast_2019"},
    {name: "CTRPv2_2015"},
    {name: "GBM_scr3"},
    {name: "GRAY_2017"},
    {name: "PRISM_2020"},
    {name: "gCSI_2019"}
]

downloadOutputs = {}

projects.forEach( (element, index) => {
    downloadOutputs[`file_${index}`] = `../../source/pharmacodb/rdata/${element.name}.rdata`
})

p = lathe.Process({
    name: "download",
    commandLine: `cwltool --outdir ../../source/pharmacodb/rdata ./download_pharmaco.cwl`,
    outputs: downloadOutputs
})
prep.Add(p)
```


# Running lathe

```
lathe run <lathe_file> <workflow_name>
```