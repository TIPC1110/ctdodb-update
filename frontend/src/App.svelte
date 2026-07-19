<script>
  import { onMount } from 'svelte';
  import { fly } from 'svelte/transition';
  import { WindowMinimise, Quit, EventsOn } from '../wailsjs/runtime/runtime.js';
  import { InstallApp, CreateDesktopShortcut, LaunchApp, SelectDirectory } from '../wailsjs/go/main/App.js';
  import logo from './assets/logo.png';

  let agreed = true;
  let keepData = true;
  let showCustomize = false;
  let showLicense = false;
  let isInstalling = false;
  let isFinished = false;
  let progress = 0;
  let currentFile = "";
  let installPath = "C:\\ctdodb";
  let lang = "vn";

  let toastMessage = "";
  let toastType = "error";
  let toastTimeout;

  function showToast(msg, type = "error") {
    toastMessage = msg;
    toastType = type;
    clearTimeout(toastTimeout);
    toastTimeout = setTimeout(() => {
      toastMessage = "";
    }, 4500);
  }

  const t = {
    en: {
      title: "Install ctdodb",
      desc: "The ultimate portable web server stack. Nginx, Apache, PHP, and MySQL.",
      overwrite: "Overwrite install",
      path: "Path:",
      browse: "Browse",
      extracting: "Extracting payload...",
      extractingProgress: "Extracting",
      keepData: "Keep data and settings of existed version",
      agreeTo: "I agree to the",
      license: "License Agreement",
      customize: "Customize",
      alertAgree: "Please agree to the License Agreement.",
      alertError: "Installation error: ",
      alertSuccess: "Installation successful! The application will start now.",
      alertUnexpected: "An error occurred: ",
      licenseTitle: "License Agreement",
      licenseText1: "Tools created by ",
      licenseText2: ", absolutely safe.",
      licenseText3: "Optimized for speed and usability.",
      licenseText4: "The team will try to update continuously:",
      licenseText5: "3-5 days / 1 minor update",
      licenseText6: "1 month / 1 major update",
      licenseText7: "Thank you all for trusting our product.",
      iAgree: "I Agree",
      finish: "Finish",
      launchApp: "Launch CTDODB"
    },
    vn: {
      title: "Cài đặt ctdodb",
      desc: "Máy chủ web di động đỉnh cao. Nginx, Apache, PHP, và MySQL.",
      overwrite: "Cài đặt đè",
      path: "Đường dẫn:",
      browse: "Duyệt",
      extracting: "Đang giải nén...",
      extractingProgress: "Đang giải nén",
      keepData: "Giữ lại dữ liệu và cấu hình của phiên bản cũ",
      agreeTo: "Tôi đồng ý với",
      license: "Điều khoản sử dụng",
      customize: "Tùy chỉnh",
      alertAgree: "Vui lòng đồng ý với Điều khoản sử dụng.",
      alertError: "Lỗi cài đặt: ",
      alertSuccess: "Cài đặt thành công! Ứng dụng sẽ khởi động ngay.",
      alertUnexpected: "Đã xảy ra lỗi: ",
      licenseTitle: "Điều khoản sử dụng",
      licenseText1: "Tools được tạo ra bởi ",
      licenseText2: ", an toàn tuyệt đối.",
      licenseText3: "Tối ưu tốc độ và độ tiện dụng.",
      licenseText4: "Team sẽ cố gắng update liên tục:",
      licenseText5: "3-5 ngày / 1 bản update nhẹ",
      licenseText6: "1 tháng / 1 bản update lớn",
      licenseText7: "Cảm ơn mọi người đã và sẽ tin dùng sản phẩm.",
      iAgree: "Tôi đồng ý",
      finish: "Hoàn tất",
      launchApp: "Khởi chạy CTDODB"
    }
  };

  $: i18n = t[lang];

  onMount(() => {
    // Listen for progress events from Go backend
    EventsOn("install:progress", (data) => {
      progress = Math.min(99, Math.floor(data.progress));
      currentFile = data.file || "";
    });
  });

  async function browsePath() {
    const dir = await SelectDirectory();
    if (dir) {
      if (dir.toLowerCase().endsWith("\\ctdodb") || dir.toLowerCase().endsWith("/ctdodb")) {
        installPath = dir;
      } else if (dir.endsWith("\\") || dir.endsWith("/")) {
        installPath = dir + "ctdodb";
      } else {
        installPath = dir + "\\ctdodb";
      }
    }
  }

  async function install() {
    if (!agreed) {
      showToast(i18n.alertAgree, "error");
      return;
    }
    isInstalling = true;
    progress = 0;
    currentFile = "";

    try {
      // 1. Extract payload
      const installRes = await InstallApp(installPath, keepData);
      if (installRes !== "success") {
        showToast(i18n.alertError + installRes, "error");
        isInstalling = false;
        progress = 0;
        currentFile = "";
        return;
      }

      // 2. Create shortcut
      await CreateDesktopShortcut(installPath);

      // Finish progress
      progress = 100;

      setTimeout(() => {
        showToast(i18n.alertSuccess, "success");
        isFinished = true;
      }, 500);

    } catch (e) {
      showToast(i18n.alertUnexpected + e, "error");
      isInstalling = false;
      progress = 0;
      currentFile = "";
    }
  }
</script>


<main class="w-screen h-screen flex flex-col bg-bg text-text-main relative overflow-hidden font-sans">
  
  <!-- Subtle Grid Background instead of wavy, looks more developer-focused -->
  <div class="absolute inset-0 opacity-[0.03] pointer-events-none" style="background-image: linear-gradient(to right, #ffffff 1px, transparent 1px), linear-gradient(to bottom, #ffffff 1px, transparent 1px); background-size: 40px 40px;"></div>
  
  <!-- Subtle Accent Glow -->
  <div class="absolute -top-40 -left-40 w-96 h-96 bg-accent rounded-full mix-blend-screen filter blur-[120px] opacity-20 pointer-events-none"></div>

  <!-- Titlebar -->
  <div class="flex items-center justify-between px-4 py-3 z-10" style="--wails-draggable: drag">
    <div class="text-muted text-xs font-semibold tracking-wider uppercase">ctdodb Setup</div>
    <div class="flex items-center gap-3" style="--wails-draggable: no-drag">
      <select bind:value={lang} class="bg-surface border border-border-color text-muted text-xs rounded px-1.5 py-0.5 outline-none hover:text-text-main focus:border-accent transition-colors cursor-pointer appearance-none">
        <option value="vn">VN</option>
        <option value="en">EN</option>
      </select>
      <button on:click={WindowMinimise} class="text-muted hover:text-text-main transition-colors focus:outline-none">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"></line></svg>
      </button>
      <button on:click={Quit} class="text-muted hover:text-danger transition-colors focus:outline-none">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
      </button>
    </div>
  </div>

  <!-- Content -->
  <div class="flex-1 flex flex-col items-center justify-center z-10">
    <!-- Logo & Branding -->
    <div class="flex flex-col items-center mb-10">
      <div class="w-28 h-28 rounded-2xl flex items-center justify-center mb-6 relative group">
        <div class="absolute inset-0 bg-accent rounded-2xl opacity-0 group-hover:opacity-20 transition-opacity blur-xl"></div>
        <img src={logo} alt="ctdodb Logo" class="w-full h-full object-contain relative z-10 drop-shadow-2xl" />
      </div>
      <h1 class="text-2xl font-bold text-text-main mb-2">{i18n.title}</h1>
      <p class="text-muted text-sm text-center max-w-xs">
        {i18n.desc}
      </p>
    </div>

    <!-- Installation Actions -->
    <div class="w-full max-w-sm flex flex-col items-center px-6">
      {#if !isInstalling && !isFinished}
        <button 
          on:click={install}
          class="w-full bg-accent hover:brightness-110 text-white font-semibold text-sm px-6 py-3.5 rounded-md transition-all shadow-[0_4px_14px_rgba(59,130,246,0.3)] active:scale-[0.98] focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2 focus:ring-offset-bg"
        >
          {i18n.overwrite}
        </button>
        <div class="mt-4 flex items-center gap-2 text-xs font-mono text-muted bg-surface-elev border border-border-color px-3 py-1.5 rounded-md w-full">
          <span>{i18n.path}</span>
          {#if showCustomize}
            <input type="text" bind:value={installPath} class="bg-bg text-text-main border border-border-color rounded px-2 py-1 outline-none focus:border-accent flex-1 min-w-0" />
            <button on:click={browsePath} class="bg-surface hover:bg-hover border border-border-color text-text-main rounded px-2 py-1 transition-colors whitespace-nowrap">{i18n.browse}</button>
          {:else}
            <span class="text-text-main truncate">{installPath}</span>
          {/if}
        </div>
      {:else if isInstalling && !isFinished}
        <div class="w-full bg-surface-elev border border-border-color rounded-lg p-5">
          <div class="flex justify-between text-sm font-medium mb-3">
            <span class="text-text-main">{i18n.extractingProgress}</span>
            <span class="text-accent font-mono">{progress}%</span>
          </div>
          <div class="h-2 w-full bg-bg rounded-full overflow-hidden mb-3">
            <div class="h-full bg-accent transition-all duration-300 ease-out shadow-[0_0_8px_rgba(59,130,246,0.5)]" style="width: {progress}%"></div>
          </div>
          {#if currentFile}
            <div class="text-xs text-muted truncate mt-2" title="{currentFile}">
              {currentFile}
            </div>
          {/if}
        </div>
      {:else if isFinished}
        <div class="w-full flex flex-col gap-3 mt-2">
          <button 
            on:click={async () => { await LaunchApp(installPath); Quit(); }}
            class="w-full bg-accent hover:brightness-110 text-white font-semibold text-sm px-6 py-3.5 rounded-md transition-all shadow-[0_4px_14px_rgba(59,130,246,0.3)] active:scale-[0.98] focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2 focus:ring-offset-bg"
          >
            {i18n.launchApp}
          </button>
          <button 
            on:click={Quit}
            class="w-full bg-surface-elev hover:bg-hover border border-border-color text-text-main font-semibold text-sm px-6 py-3.5 rounded-md transition-all active:scale-[0.98] focus:outline-none"
          >
            {i18n.finish}
          </button>
        </div>
      {/if}
    </div>
  </div>

  <!-- Bottom Options -->
  {#if !isFinished}
  <div class="w-full bg-surface border-t border-border-color px-6 py-4 z-10 flex items-center justify-between text-xs text-muted" style="--wails-draggable: no-drag">
    <div class="flex flex-col gap-2.5">
      <label class="flex items-center gap-2.5 cursor-pointer group">
        <div class="w-4 h-4 rounded-[4px] border border-border-color flex items-center justify-center transition-colors {keepData ? 'bg-accent border-accent' : 'bg-bg group-hover:border-muted'}">
          {#if keepData}
            <svg class="w-2.5 h-2.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"></path></svg>
          {/if}
        </div>
        <input type="checkbox" bind:checked={keepData} class="hidden" />
        <span class="group-hover:text-text-main transition-colors select-none">{i18n.keepData}</span>
      </label>

      <label class="flex items-center gap-2.5 cursor-pointer group">
        <div class="w-4 h-4 rounded-[4px] border border-border-color flex items-center justify-center transition-colors {agreed ? 'bg-accent border-accent' : 'bg-bg group-hover:border-muted'}">
          {#if agreed}
            <svg class="w-2.5 h-2.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"></path></svg>
          {/if}
        </div>
        <input type="checkbox" bind:checked={agreed} class="hidden" />
        <span class="group-hover:text-text-main transition-colors select-none">{i18n.agreeTo} <button type="button" class="text-accent hover:underline focus:outline-none font-medium" on:click|preventDefault|stopPropagation={() => showLicense = true}>{i18n.license}</button></span>
      </label>
    </div>

    <button on:click={() => showCustomize = !showCustomize} class="bg-surface-elev hover:bg-hover border border-border-color text-text-main px-4 py-2 rounded-md transition-colors flex items-center gap-1.5 font-medium focus:outline-none">
      {i18n.customize} 
      <svg class="w-3.5 h-3.5 transition-transform {showCustomize ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
    </button>
  </div>
  {/if}

  <!-- License Modal -->
  {#if showLicense}
    <div class="absolute inset-0 z-50 flex items-center justify-center p-6" style="--wails-draggable: no-drag">
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-bg/80 backdrop-blur-sm" on:click={() => showLicense = false}></div>
      
      <!-- Modal Content -->
      <div class="bg-surface border border-border-color rounded-2xl w-full max-w-md flex flex-col shadow-[0_20px_50px_rgba(0,0,0,0.5)] relative overflow-hidden transform transition-all">
        <!-- Subtle top glow -->
        <div class="absolute top-0 left-0 right-0 h-32 bg-accent opacity-5 blur-3xl pointer-events-none"></div>

        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-5 border-b border-border-color/50 relative z-10">
          <h2 class="text-base font-semibold text-text-main tracking-wide">{i18n.licenseTitle}</h2>
          <button on:click={() => showLicense = false} class="text-muted hover:text-text-main transition-colors focus:outline-none p-1 rounded-full hover:bg-hover">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          </button>
        </div>

        <!-- Body -->
        <div class="px-6 py-6 text-[13px] leading-relaxed text-muted space-y-4 overflow-y-auto max-h-[50vh] custom-scrollbar relative z-10">
          <p>{i18n.licenseText1}<strong class="text-text-main font-medium">ctdoteam</strong>{i18n.licenseText2}</p>
          <p>{i18n.licenseText3}</p>
          
          <div class="bg-surface-elev/50 border border-border-color/50 rounded-lg p-4 mt-2">
            <p class="font-medium text-text-main mb-2">{i18n.licenseText4}</p>
            <ul class="space-y-2">
              <li class="flex items-start gap-2">
                <span class="text-accent mt-0.5">•</span>
                <span>{i18n.licenseText5}</span>
              </li>
              <li class="flex items-start gap-2">
                <span class="text-accent mt-0.5">•</span>
                <span>{i18n.licenseText6}</span>
              </li>
            </ul>
          </div>
          
          <p class="pt-2">{i18n.licenseText7}</p>
        </div>
        
        <!-- Footer -->
        <div class="px-6 py-4 border-t border-border-color/50 bg-surface/50 flex justify-end relative z-10">
          <button on:click={() => { agreed = true; showLicense = false; }} class="bg-accent hover:brightness-110 text-white px-6 py-2 rounded-md font-medium transition-all shadow-[0_4px_14px_rgba(59,130,246,0.3)] active:scale-[0.98] focus:outline-none text-sm">
            {i18n.iAgree}
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Custom Toast Notification -->
  {#if toastMessage}
    <div transition:fly="{{ y: 30, duration: 300 }}" class="fixed bottom-6 left-1/2 -translate-x-1/2 z-[100] flex items-center gap-3 px-5 py-3 rounded-xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] border {toastType === 'error' ? 'bg-danger border-danger/30 text-white' : 'bg-surface-elev border-border-color text-text-main'}" style="--wails-draggable: no-drag">
      {#if toastType === 'error'}
        <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
      {:else}
        <svg class="w-5 h-5 flex-shrink-0 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
      {/if}
      <span class="text-sm font-medium whitespace-nowrap">{toastMessage}</span>
      <button class="ml-2 focus:outline-none" on:click={() => toastMessage = ''}>
        <svg class="w-4 h-4 opacity-60 hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
      </button>
    </div>
  {/if}
</main>

