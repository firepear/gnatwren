var url = "";
var nodes;
var metrics;
var last_seen = {};
var config = {};
var els = {};

var refresh = 120;
var past_due = 90;
var temp_warn = 80;
var temp_crit = 90;

// load up els, our list of grabbable elements
var elnames = ["doctitle", "pagetitle"];
elnames.forEach((name) => ( els[name] = document.getElementById(name) ));

function plotCPUTemps() {
    // reset cpuT.series and dimensions
    cpuT.series = [];
    cpuT.dataset.dimensions = ['timestamp'];
    cpuT.dataset.source = [];
    // create a hash for last-known temps and initazlize it to nulls
    var tl = {};
    for (let node in nodes) {
        tl[nodes[node]] = null;
        // also, while we're here, rebuild cpuT.series and dimensions
        cpuT.series.push({name: nodes[node], type: 'line', showSymbol: false, encode: {x: 'timestamp', y: nodes[node]}});
        cpuT.dataset.dimensions.push(nodes[node]);
    }

    // vivify the data we just fetched
    var temps_in = JSON.parse(this.responseText);
    // iterate over it
    for (let ts in temps_in) {
        var curData = [ts * 1000];
        for (let node in nodes) {
            var host = nodes[node];
            if (host in temps_in[ts]) {
                tl[host] = temps_in[ts][host];
                curData.push(temps_in[ts][host]);
            } else {
                curData.push(tl[host]);
            }
        }
        cpuT.dataset.source.push(curData);
    }
    // apply and render
    cpuT && cpuTChart.setOption(cpuT);
}

function getCPUTemps() {
    cpuTReq.open("GET", `${url}/cputemps-current.json`);
    cpuTReq.send();
}

function plotMachStats() {
    // wipe any existing elements
    const tbody = document.getElementById("machdata");
    while (tbody.firstChild) {
        tbody.removeChild(tbody.lastChild);
    }
    // and the node list
    nodes = [];
    // and a dict for core clock tooltips
    clocks = {};

    // get our data
    metrics = JSON.parse(this.responseText);

    for (let [host, metric] of Object.entries(metrics)) {
        // first, add this host to the list of nodes
        nodes.push(host);
        // lets calculate Avgclk for the host and build the clockspeeds tooltip
        var clktot = 0;
        var clkstr = "";
        for (let [coreid, clk] of Object.entries(metric['Cpu']['Cores'])) {
            clktot += parseFloat(clk);
            clkstr += `Core ${coreid}: ${Math.floor(clk)}MHz\n`;
        }
        clocks[host] = clkstr;
        metric['Cpu']['Avgclk'] = clktot / Object.keys(metric['Cpu']['Cores']).length;

        // now get on with constructing the document
        var tr = document.createElement("tr");
        // machine name
        var tdName = document.createElement("td");
        var nameTxt = document.createTextNode(`${host}`);
        tdName.appendChild(nameTxt);
        tdName.setAttribute('class', 'first');
        // last checkin
        var tdLast = document.createElement("td");
        var ts = Date.now() / 1000;
        last_seen[host] = metric['TS'];
        var lastTxt = document.createTextNode(`${Math.floor(ts - metric['TS'])}s`);
        tdLast.appendChild(lastTxt);
        var tsLast = new Date(metric['TS'] * 1000);
        tdLast.title = tsLast.toTimeString();
        tdLast.style.textAlign = 'right';
        if (ts - last_seen[host] > refresh + past_due) {
            tdLast.className = 'crit';
        }
        // mach desc
        var tdDesc = document.createElement("td");
        var descTxt = document.createTextNode(`${metric['Cpu']['Name']}`);
        tdDesc.appendChild(descTxt);
        // thread count
        var tdThd = document.createElement("td");
        var thdTxt = document.createTextNode(`${Object.keys(metric['Cpu']['Cores']).length}`);
        tdThd.appendChild(thdTxt);
        // cpu temp
        var tdTemp = document.createElement("td");
        var tempNum = Number(metric['Cpu']['Temp']).toFixed(2);
        var tempTxt = document.createTextNode(`${tempNum}C`);
        if (tempNum > config.ui.temp_crit_cpu) {
            tdTemp.className = 'crit';
        } else if (tempNum > config.ui.temp_hi_cpu) {
            tdTemp.className = 'warn';
        }
        tdTemp.appendChild(tempTxt);
        // cpu clock
        var tdClk = document.createElement("td");
        var clkTxt = document.createTextNode(`${Math.floor(metric['Cpu']['Avgclk'])}MHz`);
        tdClk.appendChild(clkTxt);
        tdClk.title = clocks[host];
        // uptime
        var uptime = metric['Upt'];
        var upd = Math.floor(uptime / 86400);
        uptime = uptime - upd * 86400;
        var uph = Math.floor(uptime / 3600);
        if (uph < 10) { uph = `0${uph}` }
        uptime = uptime - uph * 3600;
        var upm = Math.floor(uptime / 60);
        if (upm < 10) { upm = `0${upm}` }
        var ups = Math.floor(uptime - upm * 60);
        if (ups < 10) { ups = `0${ups}` }
        var tdUp = document.createElement("td");
        var upTxt = document.createTextNode(`${upd}d ${uph}:${upm}:${ups}`);
        tdUp.appendChild(upTxt);
        // load average
        var tdLdavg = document.createElement("td");
        var ldavgTxt = document.createTextNode(`${metric['Ldavg']}`);
        tdLdavg.appendChild(ldavgTxt);
        tdLdavg.style.textAlign = 'right';
        // memory, total
        var memtot = metric['Mem'][0] / 1024 / 1024;
        // memory, unused
        var memun = metric['Mem'][1] / metric['Mem'][0] * 100
        // memory, available
        var memav = metric['Mem'][2] / metric['Mem'][0] * 100
        var tdMemun = document.createElement("td");
        var memunTxt = document.createTextNode(`${memtot.toFixed(2)}GB : ${memun.toFixed(2)}% : ${memav.toFixed(2)}%`);
        tdMemun.appendChild(memunTxt);
        tdMemun.style.textAlign = 'right';

        // gpu desc
        var tdGdesc = document.createElement("td");
        var gdescTxt = document.createTextNode(`${metric['Gpu']['Name']}`);
        tdGdesc.appendChild(gdescTxt);
        // GPU temp
        var tdGtemp = document.createElement("td");
        var gtempTxt = document.createTextNode(`${metric['Gpu']['TempCur']}/${metric['Gpu']['TempMax']}`);
        if (metric['Gpu']['TempCur'] == "" && metric['Gpu']['TempMax'] == "") {
            gtempTxt = document.createTextNode("---")
        }
        tdGtemp.appendChild(gtempTxt);
        tdGtemp.style.textAlign = 'right';
        // GPU fan speed
        var tdGfan = document.createElement("td");
        var gfanTxt = document.createTextNode(`${metric['Gpu']['Fan']}`);
        if (metric['Gpu']['Fan'] == "") {
            gfanTxt = document.createTextNode("---")
        }
        tdGfan.appendChild(gfanTxt);
        tdGfan.style.textAlign = 'right';
        // GPU power use
        var tdGpow = document.createElement("td");
        var gpowTxt = document.createTextNode(`${metric['Gpu']['PowCur']}/${metric['Gpu']['PowMax']}`);
        if (metric['Gpu']['PowCur'] == "" && metric['Gpu']['PowMax'] == "") {
            gpowTxt = document.createTextNode("---")
        }
        tdGpow.appendChild(gpowTxt);
        tdGpow.style.textAlign = 'right';

        // add all to the row
        tr.appendChild(tdName);
        tr.appendChild(tdDesc);
        tr.appendChild(tdThd);
        tr.appendChild(tdTemp);
        tr.appendChild(tdClk);
        tr.appendChild(tdUp);
        tr.appendChild(tdLdavg);
        tr.appendChild(tdMemun);
        tr.appendChild(tdGdesc);
        tr.appendChild(tdGtemp);
        tr.appendChild(tdGfan);
        tr.appendChild(tdGpow);
        tr.appendChild(tdLast);
        tbody.appendChild(tr);
    }
    // sort the nodes list
    nodes.sort();
    // scan for nodes that have fallen out of the current table
    for (let [host, lsts] of Object.entries(last_seen)) {
        if (nodes.includes(host)) {
            continue;
        }
        var tr = document.createElement("tr");
        var td = document.createElement("td");
        var ts = Date.now() / 1000;         // div by 1k to get unix time
        var lsdate = new Date(lsts * 1000); // multiply by 1k to get a valid Date obj
        var tdTxt = document.createTextNode(`${host} missing; last seen ${Math.floor(ts - lsts)}s ago (${lsdate.toString()})`);
        td.colSpan = 13;
        td.className = 'warn';
        td.appendChild(tdTxt);
        tr.appendChild(td);
        tbody.appendChild(tr);
    }

    // and now that System Data table is populated, we can draw graphs
    getCPUTemps();
}

function getMachStats() {
    machSReq.open("GET", `${url}/overview.json`);
    machSReq.send();
}

async function getConf() {
    config = await fetch(`${url}/gwgather-config.json`).then((r) => {return r.json() });
    els["doctitle"].replaceChildren();
    els["doctitle"].insertAdjacentHTML("beforeend", config.ui.title);
    els["pagetitle"].replaceChildren();
    els["pagetitle"].insertAdjacentHTML("beforeend", config.ui.title);

    machSReq.addEventListener("load", plotMachStats);
    cpuTReq.addEventListener("load", plotCPUTemps);
}

// ////////////////////////////////////////////////////////////// main routine

getConf();

// one of these for every chart
var cpuTDom = document.getElementById('cputemps');
var cpuTChart = echarts.init(cpuTDom);
var cpuT;
// set options
cpuT = {
    title: { text: 'CPU temperatures, last hour' },
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'time', boundaryGap: ['20%', '20%'], min: 'dataMin', max: 'dataMax' },
    yAxis: { type: 'value', min: 'dataMin' },
    series: [],
    dataset: {
        source: [],
        dimensions: ['timestamp']
    }
};
/*
  var cpuADom = document.getElementById('cpuavg');
  var cpuAChart = echarts.init(cpuADom);
  var cpuA;
  // set options
  cpuA = {
  title: { text: 'CPU avg clock, last hour' },
  tooltip: { trigger: 'axis' },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  xAxis: { type: 'time', boundaryGap: ['20%', '20%'], min: 'dataMin', max: 'dataMax' },
  yAxis: { type: 'value', min: 'dataMin' },
  series: [],
  dataset: {
  source: [],
  dimensions: ['timestamp']
  }
  };
*/

var machSReq = new XMLHttpRequest();
var cpuTReq = new XMLHttpRequest();
//var cpuAReq = new XMLHttpRequest();
//cpuAReq.addEventListener("load", plotCPUAvgs);

getMachStats();
setInterval(getMachStats, 120000);
