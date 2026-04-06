const PAGE_TITLES = {
  overview: "总览",
  hardware: "硬件信息",
  software: "软件信息",
  advanced: "进阶诊断",
  raw: "原始 JSON"
};

const KEY_LABELS = {
  cores_logical: "逻辑核心数",
  cores_physical: "物理核心数",
  model: "型号",
  display_name: "系统名称",
  version: "系统版本",
  build: "构建号",
  major: "主版本号",
  minor: "次版本号",
  service_pack_major: "服务包版本",
  system_dir: "系统目录",
  computer_name: "计算机名",
  user_name: "当前用户",
  os: "操作系统",
  name: "名称",
  mac: "MAC 地址",
  ipv4: "IPv4 地址",
  model: "型号",
  manufacturer: "厂商",
  part_number: "部件号",
  capacity_gb: "容量(GB)",
  configured_clock_mhz: "频率(MHz)",
  status: "状态",
  link_speed: "链路速率",
  network_physical_adapters: "物理网卡",
  modules: "内存条",
  activation: "激活状态",
  startup_items: "启动项",
  battery: "电池信息",
  bios: "BIOS 信息",
  baseboard: "主板信息",
  gpu: "显卡信息",
  boot_mode: "启动模式",
  last_boot_time: "最近启动时间",
  uptime_seconds: "已运行秒数",
  recent_bugcheck: "最近蓝屏记录",
  adapter: "网络配置",
  defaultipgateway: "默认网关",
  dnsserversearchorder: "DNS 服务器",
  ipaddress: "IP 地址",
  licensestatus: "许可状态",
  description: "描述",
  leveldisplayname: "级别",
  providername: "提供者",
  timecreated: "时间",
  id: "ID",
  command: "命令",
  location: "位置",
  estimatedchargeremaining: "剩余电量(%)",
  batterystatus: "电池状态",
  designcapacity: "设计容量",
  fullchargecapacity: "满充容量",
  cyclecount: "循环次数",
  smbiosbiosversion: "BIOS 版本",
  releasedate: "发布时间",
  serialnumber: "序列号",
  product: "产品型号",
  displayname: "名称",
  displayversion: "版本",
  publisher: "发布者",
  installdate: "安装日期",
  key_drivers: "关键驱动",
  predict_status: "预测状态",
  vendor_data: "厂商数据",
  physical_disks: "物理磁盘",
  volumes: "卷信息",
  healthstatus: "健康状态",
  operationalstatus: "运行状态",
  mediatype: "介质类型",
  size: "容量",
  sizeremaining: "剩余容量",
  driveletter: "盘符",
  filesystemlabel: "卷标",
  problematic_devices: "异常设备",
  configmanagererrorcode: "错误码",
  pnpclass: "设备类别",
  browsers: "浏览器",
  installed_software: "已安装软件",
  drive: "盘符",
  partition_number: "分区号",
  filesystem: "文件系统",
  total_gb: "总容量(GB)",
  used_gb: "已用(GB)",
  free_gb: "可用(GB)",
  disk_number: "物理盘号",
  disk_type: "磁盘类型",
  bus_type: "总线类型",
  memory_load_percent: "内存占用(%)",
  available_gb: "可用内存(GB)",
  used_gb_mem: "已用内存(GB)",
  total_gb_mem: "总内存(GB)",
  problematic_devices: "异常设备",
  has_peripheral_issues: "外设是否异常",
  peripheral_issues: "外设异常详情",
  ping_gateway_ok: "网关连通",
  ping_external_ok: "外网连通",
  dns_ok: "DNS 可用",
  winhttp_proxy_enabled: "系统代理开启",
  cpu_percent: "CPU 占用(%)",
  disk_busy_percent: "磁盘繁忙(%)",
  memory_available_mb: "可用内存(MB)",
  top_cpu_process_names: "高 CPU 进程",
  top_memory_process_names: "高内存进程"
};

const TOOL_KEYS = ["go", "node", "python", "java", "git", "docker", "kubectl", "dotnet"];

let report = null;
let currentPage = "overview";

function label(k) {
  const key = String(k || "");
  return KEY_LABELS[key] || KEY_LABELS[key.toLowerCase()] || key;
}
function safe(v, d = "-") { return v === null || v === undefined || v === "" ? d : v; }
function escapeHtml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}
function linesHtml(items) {
  const arr = (items || []).map((x) => String(x || "").trim()).filter(Boolean);
  if (arr.length === 0) return "-";
  return arr.map((x) => `<div>${escapeHtml(x)}</div>`).join("");
}
function gb(v) {
  const n = Number(v);
  if (Number.isNaN(n)) return "-";
  return n.toFixed(1);
}

function gpuRank(name) {
  const n = String(name || "").toLowerCase();
  let score = 0;
  if (n.includes("amd") || n.includes("radeon")) score += 30;
  if (n.includes("nvidia") || n.includes("geforce")) score += 25;
  if (n.includes("intel")) score += 10;
  if (n.includes("virtual") || n.includes("gameviewer") || n.includes("remote") || n.includes("basic render")) score -= 40;
  return score;
}

function getSortedGpuList(gpu) {
  const arr = Array.isArray(gpu) ? gpu.slice() : (gpu ? [gpu] : []);
  arr.sort((a, b) => gpuRank(b?.Name || b?.name) - gpuRank(a?.Name || a?.name));
  return arr;
}

function pct(used, total) {
  const u = Number(used);
  const t = Number(total);
  if (Number.isNaN(u) || Number.isNaN(t) || t <= 0) return 0;
  return Math.max(0, Math.min(100, (u / t) * 100));
}

function ringColorByPct(p) {
  if (p > 75) return "#cf3e36"; // red
  if (p > 25) return "#2a6fdb"; // blue
  return "#1f8a3f"; // green
}

function renderRing(title, usedGb, totalGb, color = null) {
  const p = pct(usedGb, totalGb);
  const c = color || ringColorByPct(p);
  const freeGb = Number(totalGb) - Number(usedGb);
  const card = document.createElement("div");
  card.className = "ring-card";
  card.innerHTML = `
    <div class="donut" style="--pct:${p.toFixed(1)};--c1:${c}">
      <div class="donut-label">${p.toFixed(1)}%</div>
    </div>
    <div class="ring-info">
      <div class="title">${title}</div>
      <div class="line">总量：${gb(totalGb)} GB</div>
      <div class="line">已用：${gb(usedGb)} GB</div>
      <div class="line">未用：${gb(freeGb)} GB</div>
    </div>
  `;
  return card;
}

function createPanel(title) {
  const p = document.createElement("section");
  p.className = "panel";
  p.innerHTML = `<h3>${title}</h3><div class="panel-body"></div>`;
  return p;
}

function addKV(body, k, v, html = false) {
  const tpl = document.getElementById("kvTpl");
  const node = tpl.content.cloneNode(true);
  node.querySelector(".k").textContent = k;
  const ve = node.querySelector(".v");
  if (html) ve.innerHTML = v;
  else ve.textContent = String(safe(v));
  body.appendChild(node);
}

function statusBadge(installed) {
  if (installed) return '<span class="status ok">已安装</span>';
  return '<span class="status no">未安装</span>';
}

function boolCN(v) { return v ? "是" : "否"; }

function renderOverview(root) {
  const hw = report.hardware || {};
  const adv = report.advanced || {};
  const cpu = hw.cpu || {};
  const mem = hw.memory || {};
  const net = Array.isArray(hw.network) ? hw.network : [];
  const physicalNics = Array.isArray(hw.network_physical_adapters) ? hw.network_physical_adapters : [];
  const pdisks = Array.isArray(hw.disk?.physical_disks) ? hw.disk.physical_disks : [];
  const memModules = Array.isArray(mem.modules) ? mem.modules : [];
  const hh = adv.hardware_health || {};
  const gpu = hh.gpu;
  const gpuList = getSortedGpuList(gpu);
  const bios = hh.bios || {};
  const baseboard = hh.baseboard || {};
  const p1 = createPanel("关键指标");
  const b1 = p1.querySelector(".panel-body");
  const tw = document.createElement("div");
  tw.className = "table-wrap";
  const table = document.createElement("table");
  table.innerHTML = `<thead><tr><th>项目</th><th>型号/名称</th><th>补充信息</th></tr></thead><tbody></tbody>`;
  const tb = table.querySelector("tbody");

  const cpuFreq = cpu.frequency_mhz ? `${cpu.frequency_mhz} MHz` : "-";
  const trCPU = document.createElement("tr");
  trCPU.innerHTML = `<td>CPU</td><td>${safe(cpu.model)}</td><td>频率：${cpuFreq}</td>`;
  tb.appendChild(trCPU);

  const gpuNames = gpuList.map((g) => safe(g.Name || g.name)).filter(Boolean);
  const gpuVRams = gpuList.map((g) => {
    const raw = g.AdapterRAM ?? g.adapterram;
    const n = Number(raw);
    if (!Number.isNaN(n) && n > 0) {
      return `${(n / (1024 * 1024 * 1024)).toFixed(1)} GB`;
    }
    return "-";
  });
  const trGPU = document.createElement("tr");
  trGPU.innerHTML = `<td>GPU</td><td>${linesHtml(gpuNames)}</td><td>${linesHtml(gpuVRams.map((v) => `显存：${v}`))}</td>`;
  tb.appendChild(trGPU);

  const memNames = memModules.map((m) => `${safe(m.manufacturer)} ${safe(m.part_number)}`).filter(Boolean);
  const memSizes = memModules.map((m) => `${gb(m.capacity_gb)} GB`);
  const trMem = document.createElement("tr");
  trMem.innerHTML = `<td>内存条</td><td>${linesHtml(memNames)}</td><td>${linesHtml(memSizes.map((s) => `容量：${s}`))}</td>`;
  tb.appendChild(trMem);

  const diskNames = pdisks.map((d) => `${safe(d.model)} (${safe(d.disk_type)})`).filter(Boolean);
  const diskSizes = pdisks.map((d) => `${gb(d.total_gb)} GB`);
  const trDisk = document.createElement("tr");
  trDisk.innerHTML = `<td>硬盘</td><td>${linesHtml(diskNames)}</td><td>${linesHtml(diskSizes.map((s) => `容量：${s}`))}</td>`;
  tb.appendChild(trDisk);

  const biosVer = safe(bios.SMBIOSBIOSVersion || bios.smbiosbiosversion);
  const biosBrand = safe(bios.Manufacturer || bios.manufacturer);
  const trBios = document.createElement("tr");
  trBios.innerHTML = `<td>BIOS</td><td>${biosVer}</td><td>品牌：${biosBrand}</td>`;
  tb.appendChild(trBios);

  const boardModel = safe(baseboard.Product || baseboard.product);
  const boardBrand = safe(baseboard.Manufacturer || baseboard.manufacturer);
  const trBoard = document.createElement("tr");
  trBoard.innerHTML = `<td>主板</td><td>${boardModel}</td><td>品牌：${boardBrand}</td>`;
  tb.appendChild(trBoard);

  const physicalSet = new Set(physicalNics.map((x) => String(x.name || "").toLowerCase()));
  const physicalActive = net.filter((n) => physicalSet.has(String(n.name || "").toLowerCase()));
  const virtualActive = net.filter((n) => !physicalSet.has(String(n.name || "").toLowerCase()));
  if (physicalActive.length > 0 || virtualActive.length > 0) {
    const netNameLines = [];
    const netModelLines = [];

    if (physicalActive.length > 0) {
      physicalActive.forEach((n) => netNameLines.push(`- ${safe(n.name)}`));
      physicalActive.forEach((n) => {
        const p = physicalNics.find((x) => String(x.name || "").toLowerCase() === String(n.name || "").toLowerCase());
        netModelLines.push(`- 型号：${safe(p?.model)}`);
      });
    }
    if (virtualActive.length > 0) {
      virtualActive.forEach((n) => netNameLines.push(`- ${safe(n.name)}`));
      virtualActive.forEach((n) => netModelLines.push("- 型号：虚拟/隧道/代理网卡"));
    }

    const trNic = document.createElement("tr");
    trNic.innerHTML = `<td>网卡</td><td>${linesHtml(netNameLines)}</td><td>${linesHtml(netModelLines)}</td>`;
    tb.appendChild(trNic);
  }

  tw.appendChild(table);
  b1.appendChild(tw);

  const chartsPanel = createPanel("容量占用可视化");
  const cb = chartsPanel.querySelector(".panel-body");
  const ringGrid = document.createElement("div");
  ringGrid.className = "ring-grid";

  const logical = Array.isArray(hw.disk?.logical_partitions) ? hw.disk.logical_partitions : [];
  logical.forEach((d) => {
    const drive = safe(d.drive);
    ringGrid.appendChild(
      renderRing(`${drive} 盘占用`, Number(d.used_gb) || 0, Number(d.total_gb) || 0)
    );
  });

  ringGrid.appendChild(renderRing("物理内存占用", Number(mem.used_gb) || 0, Number(mem.total_gb) || 0));
  ringGrid.appendChild(
    renderRing("虚拟内存占用", Number(mem.virtual_used_gb) || 0, Number(mem.virtual_total_gb) || 0)
  );
  const perf = adv.performance_diagnostics || {};
  if (Number(perf.gpu_memory_total_mb) > 0) {
    ringGrid.appendChild(
      renderRing(
        "显存占用",
        Number(perf.gpu_memory_used_mb || 0) / 1024,
        Number(perf.gpu_memory_total_mb || 0) / 1024
      )
    );
  }
  cb.appendChild(ringGrid);
  root.appendChild(p1);
  root.appendChild(chartsPanel);
}

function renderStorageTopology(parent) {
  const panel = createPanel("存储拓扑（按物理磁盘）");
  const body = panel.querySelector(".panel-body");
  const disks = Array.isArray(report.hardware?.disk?.physical_disks) ? report.hardware.disk.physical_disks : [];

  if (disks.length === 0) {
    const e = document.createElement("div");
    e.className = "error";
    e.textContent = "未获取到物理磁盘信息";
    body.appendChild(e);
    parent.appendChild(panel);
    return;
  }

  disks.forEach((d) => {
    const wrap = document.createElement("div");
    wrap.className = "disk-card";
    const head = document.createElement("div");
    head.className = "disk-head";
    head.innerHTML = `
      <div><div class="muted">物理盘</div><div>#${safe(d.disk_number)}</div></div>
      <div><div class="muted">型号</div><div>${safe(d.model)}</div></div>
      <div><div class="muted">类型</div><div><span class="tag">${safe(d.disk_type)}</span></div></div>
      <div><div class="muted">总容量</div><div>${gb(d.total_gb)} GB</div></div>
    `;
    wrap.appendChild(head);

    const tw = document.createElement("div");
    tw.className = "table-wrap";
    const table = document.createElement("table");
    table.innerHTML = `
      <thead><tr><th>盘符</th><th>分区号</th><th>文件系统</th><th>总容量(GB)</th><th>已用(GB)</th><th>可用(GB)</th></tr></thead>
      <tbody></tbody>
    `;
    const tbody = table.querySelector("tbody");
    const parts = Array.isArray(d.logical_partitions) ? d.logical_partitions : [];
    parts.forEach((p) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${safe(p.drive)}</td><td>${safe(p.partition_number)}</td><td>${safe(p.filesystem)}</td><td>${gb(p.total_gb)}</td><td>${gb(p.used_gb)}</td><td>${gb(p.free_gb)}</td>`;
      tbody.appendChild(tr);
    });
    tw.appendChild(table);
    wrap.appendChild(tw);
    body.appendChild(wrap);
  });
  parent.appendChild(panel);
}

function renderHardware(root) {
  const hw = report.hardware || {};
  const adv = report.advanced || {};
  const cpu = hw.cpu || {};
  const mem = hw.memory || {};
  const os = hw.os || {};
  const net = Array.isArray(hw.network) ? hw.network : [];

  const pCPU = createPanel("处理器");
  const bCPU = pCPU.querySelector(".panel-body");
  addKV(bCPU, "型号", safe(cpu.model));
  addKV(bCPU, "逻辑核心数", safe(cpu.cores_logical));
  addKV(bCPU, "物理核心数", safe(cpu.cores_physical));
  root.appendChild(pCPU);

  const pMem = createPanel("内存");
  const bMem = pMem.querySelector(".panel-body");
  addKV(bMem, "总内存(GB)", `${gb(mem.total_gb)} GB`);
  addKV(bMem, "可用内存(GB)", `${gb(mem.available_gb)} GB`);
  addKV(bMem, "已用内存(GB)", `${gb(mem.used_gb)} GB`);
  addKV(bMem, "内存占用(%)", safe(mem.memory_load_percent));
  const mods = Array.isArray(mem.modules) ? mem.modules : [];
  if (mods.length > 0) {
    mods.forEach((m, i) => {
      addKV(
        bMem,
        `内存条${i + 1}`,
        `${safe(m.manufacturer)} ${safe(m.part_number)} ${gb(m.capacity_gb)}GB ${safe(m.configured_clock_mhz, "-")}MHz`
      );
    });
  }
  root.appendChild(pMem);

  const pNet = createPanel("网络适配器");
  const bNet = pNet.querySelector(".panel-body");
  const pnics = Array.isArray(hw.network_physical_adapters) ? hw.network_physical_adapters : [];
  if (pnics.length === 0 && net.length === 0) {
    addKV(bNet, "网卡", "无");
  } else {
    const tw = document.createElement("div");
    tw.className = "table-wrap";
    const table = document.createElement("table");
    table.innerHTML = `<thead><tr><th>网卡名称</th><th>网卡型号</th><th>IPv4</th><th>MAC</th><th>状态</th></tr></thead><tbody></tbody>`;
    const tbody = table.querySelector("tbody");
    const pMap = new Map();
    pnics.forEach((x) => pMap.set(String(x.name || "").toLowerCase(), x));

    const physicalActive = net.filter((n) => pMap.has(String(n.name || "").toLowerCase()));
    const virtualActive = net.filter((n) => !pMap.has(String(n.name || "").toLowerCase()));
    const ordered = [...physicalActive, ...virtualActive];

    ordered.forEach((n) => {
      const p = pMap.get(String(n.name || "").toLowerCase());
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${safe(n.name)}</td><td>${safe(p?.model)}</td><td>${safe(n.ipv4)}</td><td>${safe(n.mac)}</td><td>${safe(p?.status, "-")}</td>`;
      tbody.appendChild(tr);
    });
    pnics.forEach((p) => {
      const hasShown = net.some((n) => String(n.name || "").toLowerCase() === String(p.name || "").toLowerCase());
      if (!hasShown) {
        const tr = document.createElement("tr");
        tr.innerHTML = `<td>${safe(p.name)}</td><td>${safe(p.model)}</td><td>-</td><td>${safe(p.mac)}</td><td>${safe(p.status, "-")}</td>`;
        tbody.appendChild(tr);
      }
    });
    tw.appendChild(table);
    bNet.appendChild(tw);
  }
  root.appendChild(pNet);

  const pOS = createPanel("操作系统");
  const bOS = pOS.querySelector(".panel-body");
  ["display_name", "version", "build", "system_dir", "computer_name", "user_name"].forEach((k) => {
    addKV(bOS, label(k), safe(os[k]));
  });
  root.appendChild(pOS);

  const pGPU = createPanel("显卡信息");
  const bGPU = pGPU.querySelector(".panel-body");
  const gpu = adv.hardware_health?.gpu;
  const gpuList = getSortedGpuList(gpu);
  if (gpuList.length === 0) {
    addKV(bGPU, "显卡", "无");
  } else {
    const lines = gpuList.map((g) => `${safe(g.Name || g.name)} / 驱动 ${safe(g.DriverVersion || g.driverversion)} / 状态 ${safe(g.Status || g.status)}`);
    addKV(bGPU, "显卡", linesHtml(lines), true);
  }
  root.appendChild(pGPU);

  renderStorageTopology(root);
}

function renderSoftware(root) {
  const sw = report.software || {};
  const adv = report.advanced || {};

  const pTools = createPanel("开发工具安装状态");
  const bTools = pTools.querySelector(".panel-body");
  const tw = document.createElement("div");
  tw.className = "table-wrap";
  const table = document.createElement("table");
  table.innerHTML = `<thead><tr><th>工具</th><th>安装状态</th><th>版本/信息</th></tr></thead><tbody></tbody>`;
  const tbody = table.querySelector("tbody");
  TOOL_KEYS.forEach((k) => {
    const t = sw[k] || {};
    const tr = document.createElement("tr");
    const installed = !!t.installed;
    const ver = t.version || t.version_raw || t.raw || "-";
    tr.innerHTML = `<td>${k.toUpperCase()}</td><td>${statusBadge(installed)}</td><td>${safe(ver)}</td>`;
    tbody.appendChild(tr);
  });
  tw.appendChild(table);
  bTools.appendChild(tw);
  root.appendChild(pTools);

  const pInstalled = createPanel("已安装软件");
  const bInstalled = pInstalled.querySelector(".panel-body");
  const inv = adv.software_inventory || {};
  const softwares = Array.isArray(inv.installed_software) ? inv.installed_software : [];
  if (softwares.length === 0) {
    addKV(bInstalled, "已安装软件", "无");
  } else {
    const tw2 = document.createElement("div");
    tw2.className = "table-wrap";
    const table2 = document.createElement("table");
    table2.innerHTML = `<thead><tr><th>名称</th><th>版本</th><th>发布者</th></tr></thead><tbody></tbody>`;
    const tbody2 = table2.querySelector("tbody");
    softwares.forEach((s) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${safe(s.DisplayName || s.displayname)}</td><td>${safe(s.DisplayVersion || s.displayversion)}</td><td>${safe(s.Publisher || s.publisher)}</td>`;
      tbody2.appendChild(tr);
    });
    tw2.appendChild(table2);
    bInstalled.appendChild(tw2);
  }
  root.appendChild(pInstalled);
}

function renderAdvancedValue(container, key, value) {
  if (value === null || value === undefined) {
    addKV(container, label(key), "-");
    return;
  }

  if (Array.isArray(value)) {
    if (value.length === 0) {
      addKV(container, label(key), "空");
      return;
    }
    const block = document.createElement("div");
    block.className = "sub-block";
    block.innerHTML = `<div class="sub-title">${label(key)}</div><div class="sub-body"></div>`;
    const body = block.querySelector(".sub-body");

    value.forEach((item, idx) => {
      const itemBlock = document.createElement("div");
      itemBlock.className = "sub-block";
      itemBlock.innerHTML = `<div class="sub-title">第 ${idx + 1} 项</div><div class="sub-body"></div>`;
      const ib = itemBlock.querySelector(".sub-body");
      if (item && typeof item === "object") {
        Object.entries(item).forEach(([k, v]) => renderAdvancedValue(ib, k, v));
      } else {
        addKV(ib, "值", item);
      }
      body.appendChild(itemBlock);
    });
    container.appendChild(block);
    return;
  }

  if (typeof value === "object") {
    const block = document.createElement("div");
    block.className = "sub-block";
    block.innerHTML = `<div class="sub-title">${label(key)}</div><div class="sub-body"></div>`;
    const body = block.querySelector(".sub-body");
    Object.entries(value).forEach(([k, v]) => renderAdvancedValue(body, k, v));
    container.appendChild(block);
    return;
  }

  addKV(container, label(key), typeof value === "boolean" ? boolCN(value) : value);
}

function renderAdvanced(root) {
  const adv = report.advanced || {};

  const pHard = createPanel("硬件健康（精简）");
  const bHard = pHard.querySelector(".panel-body");
  const hh = adv.hardware_health || {};
  ["battery", "bios", "baseboard"].forEach((k) => renderAdvancedValue(bHard, k, hh[k]));
  root.appendChild(pHard);

  const pSys = createPanel("系统诊断");
  const bSys = pSys.querySelector(".panel-body");
  const sys = adv.system_diagnostics || {};
  addKV(bSys, "启动模式", safe(sys.boot_mode));
  addKV(bSys, "已运行秒数", safe(sys.uptime_seconds));
  addKV(bSys, "激活状态", safe(sys.activation?.LicenseStatus ?? sys.activation?.licensestatus));
  const startup = Array.isArray(sys.startup_items) ? sys.startup_items : [];
  if (startup.length > 0) {
    const tw = document.createElement("div");
    tw.className = "table-wrap";
    const table = document.createElement("table");
    table.innerHTML = `<thead><tr><th>名称</th><th>命令</th><th>位置</th><th>用户</th></tr></thead><tbody></tbody>`;
    const tb = table.querySelector("tbody");
    startup.forEach((s) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${safe(s.Name || s.name)}</td><td>${safe(s.Command || s.command)}</td><td>${safe(s.Location || s.location)}</td><td>${safe(s.User || s.user)}</td>`;
      tb.appendChild(tr);
    });
    tw.appendChild(table);
    bSys.appendChild(tw);
  }
  root.appendChild(pSys);

  const pNet = createPanel("网络诊断");
  const bNet = pNet.querySelector(".panel-body");
  const nd = adv.network_diagnostics || {};
  const adapter = nd.adapter || {};
  addKV(bNet, "网关连通", boolCN(!!nd.ping_gateway_ok));
  addKV(bNet, "外网连通", boolCN(!!nd.ping_external_ok));
  addKV(bNet, "DNS 可用", boolCN(!!nd.dns_ok));
  addKV(bNet, "系统代理开启", boolCN(!!nd.winhttp_proxy_enabled));
  const twN = document.createElement("div");
  twN.className = "table-wrap";
  const tN = document.createElement("table");
  const dns = Array.isArray(adapter.DNSServerSearchOrder || adapter.dnsserversearchorder)
    ? (adapter.DNSServerSearchOrder || adapter.dnsserversearchorder).join(" / ")
    : safe(adapter.DNSServerSearchOrder || adapter.dnsserversearchorder);
  const ips = Array.isArray(adapter.IPAddress || adapter.ipaddress)
    ? (adapter.IPAddress || adapter.ipaddress).join(" / ")
    : safe(adapter.IPAddress || adapter.ipaddress);
  const gws = Array.isArray(adapter.DefaultIPGateway || adapter.defaultipgateway)
    ? (adapter.DefaultIPGateway || adapter.defaultipgateway).join(" / ")
    : safe(adapter.DefaultIPGateway || adapter.defaultipgateway);
  tN.innerHTML = `<thead><tr><th>网卡描述</th><th>IP</th><th>网关</th><th>DNS</th></tr></thead>
    <tbody><tr><td>${safe(adapter.Description || adapter.description)}</td><td>${ips}</td><td>${gws}</td><td>${dns}</td></tr></tbody>`;
  twN.appendChild(tN);
  bNet.appendChild(twN);
  root.appendChild(pNet);

  const pPerf = createPanel("性能诊断");
  const bPerf = pPerf.querySelector(".panel-body");
  const perf = adv.performance_diagnostics || {};
  addKV(bPerf, "CPU 占用(%)", safe(perf.cpu_percent));
  addKV(bPerf, "磁盘繁忙(%)", safe(perf.disk_busy_percent));
  addKV(bPerf, "可用内存(MB)", safe(perf.memory_available_mb));
  const chips1 = document.createElement("div");
  chips1.className = "chip-row";
  (Array.isArray(perf.top_cpu_process_names) ? perf.top_cpu_process_names : []).forEach((n) => {
    const c = document.createElement("span");
    c.className = "chip";
    c.textContent = String(n);
    chips1.appendChild(c);
  });
  if (chips1.childElementCount > 0) {
    addKV(bPerf, "高 CPU 进程", "");
    bPerf.lastElementChild.querySelector(".v").appendChild(chips1);
  }
  const chips2 = document.createElement("div");
  chips2.className = "chip-row";
  (Array.isArray(perf.top_memory_process_names) ? perf.top_memory_process_names : []).forEach((n) => {
    const c = document.createElement("span");
    c.className = "chip";
    c.textContent = String(n);
    chips2.appendChild(c);
  });
  if (chips2.childElementCount > 0) {
    addKV(bPerf, "高内存进程", "");
    bPerf.lastElementChild.querySelector(".v").appendChild(chips2);
  }
  root.appendChild(pPerf);

  const pDrv = createPanel("驱动诊断");
  const bDrv = pDrv.querySelector(".panel-body");
  renderAdvancedValue(bDrv, "problematic_devices", adv.driver_diagnostics?.problematic_devices);
  root.appendChild(pDrv);
}

function renderRaw(root) {
  const panel = createPanel("原始 JSON");
  const pre = document.createElement("pre");
  pre.className = "json";
  pre.textContent = JSON.stringify(report.original, null, 2);
  panel.querySelector(".panel-body").appendChild(pre);
  root.appendChild(panel);
}

function renderPage() {
  document.getElementById("pageTitle").textContent = PAGE_TITLES[currentPage] || "详情";
  const body = document.getElementById("pageBody");
  body.innerHTML = "";

  if (currentPage === "overview") renderOverview(body);
  else if (currentPage === "hardware") renderHardware(body);
  else if (currentPage === "software") renderSoftware(body);
  else if (currentPage === "advanced") renderAdvanced(body);
  else renderRaw(body);

  // 切页后回到顶部，避免沿用上一页滚动位置
  body.scrollTop = 0;
  const content = document.querySelector(".content");
  if (content) content.scrollTop = 0;
  window.scrollTo({ top: 0, behavior: "auto" });
}

async function loadReport() {
  const res = await fetch("/api/report", { cache: "no-store" });
  if (!res.ok) throw new Error("接口请求失败");
  const data = await res.json();
  applyReportData(data);
}

function applyReportData(data) {
  report = {
    original: data,
    hostname: data.hostname,
    runtime: data.runtime,
    timestamp: data.timestamp,
    hardware: data.hardware || {},
    software: data.software || {},
    advanced: data.advanced || {}
  };

  const t = data.timestamp ? new Date(data.timestamp * 1000).toLocaleString("zh-CN") : "-";
  document.getElementById("meta").textContent = `主机：${safe(data.hostname)} | 运行时：${safe(data.runtime)} | 采集时间：${t}`;
  document.getElementById("refreshStatus").textContent = `数据状态：已刷新（${new Date().toLocaleTimeString("zh-CN", { hour12: false })}）`;
  renderPage();
}

function bindEvents() {
  const nav = document.getElementById("nav");
  nav.querySelectorAll("button").forEach((btn) => {
    btn.addEventListener("click", () => {
      currentPage = btn.dataset.page;
      nav.querySelectorAll("button").forEach((x) => x.classList.remove("active"));
      btn.classList.add("active");
      renderPage();
    });
  });

  document.getElementById("reloadBtn").addEventListener("click", () => window.location.reload());
  document.getElementById("refreshBtn").addEventListener("click", async () => {
    const btn = document.getElementById("refreshBtn");
    const status = document.getElementById("refreshStatus");
    btn.disabled = true;
    btn.textContent = "刷新中...";
    status.textContent = "数据状态：正在刷新...";
    try {
      const res = await fetch("/api/refresh", { method: "POST", cache: "no-store" });
      if (!res.ok) throw new Error("刷新请求失败");
      const data = await res.json();
      applyReportData(data);
    } catch (err) {
      document.getElementById("meta").textContent = `加载失败：${err.message}`;
      status.textContent = "数据状态：刷新失败";
    } finally {
      btn.disabled = false;
      btn.textContent = "刷新数据";
    }
  });
}

window.addEventListener("DOMContentLoaded", async () => {
  bindEvents();
  try {
    await loadReport();
  } catch (err) {
    document.getElementById("meta").textContent = `加载失败：${err.message}`;
  }
});
