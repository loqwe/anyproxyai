<template>
  <n-config-provider :theme="isDark ? darkTheme : null" :theme-overrides="themeOverrides">
    <n-layout style="height: 100vh;">
      <!-- Top Bar -->
      <n-layout-header bordered style="height: 64px; padding: 0 24px; display: flex; align-items: center; justify-content: space-between;">
        <div style="display: flex; align-items: center; gap: 12px;">
          <n-icon size="32" color="#18a058">
            <ServerIcon />
          </n-icon>
          <span style="font-size: 20px; font-weight: 600;">AnyProxyAi</span>
        </div>

        <!-- Navigation Tabs -->
        <div style="display: flex; align-items: center; gap: 8px;">
          <n-button
            size="small"
            :type="currentPage === 'home' ? 'primary' : 'default'"
            :ghost="currentPage !== 'home'"
            @click="currentPage = 'home'"
          >
            <template #icon>
              <n-icon><HomeIcon /></n-icon>
            </template>
            {{ t('nav.home') }}
          </n-button>

          <n-button
            size="small"
            :type="currentPage === 'models' ? 'primary' : 'default'"
            :ghost="currentPage !== 'models'"
            @click="currentPage = 'models'"
          >
            <template #icon>
              <n-icon><ListIcon /></n-icon>
            </template>
            {{ t('nav.models') }}
          </n-button>

          <n-button
            size="small"
            :type="currentPage === 'stats' ? 'primary' : 'default'"
            :ghost="currentPage !== 'stats'"
            @click="currentPage = 'stats'"
          >
            <template #icon>
              <n-icon><BarChartIcon /></n-icon>
            </template>
            {{ t('nav.stats') }}
          </n-button>

          <n-button
            size="small"
            :type="currentPage === 'logs' ? 'primary' : 'default'"
            :ghost="currentPage !== 'logs'"
            @click="currentPage = 'logs'"
          >
            <template #icon>
              <n-icon><DocumentTextIcon /></n-icon>
            </template>
            {{ t('nav.logs') }}
          </n-button>

          <n-button
            size="small"
            :type="currentPage === 'health' ? 'primary' : 'default'"
            :ghost="currentPage !== 'health'"
            @click="currentPage = 'health'; loadHealthStatus()"
          >
            <template #icon>
              <n-icon><PulseIcon /></n-icon>
            </template>
            {{ t('nav.health') }}
          </n-button>

          <n-button
            size="small"
            :type="currentPage === 'traces' ? 'primary' : 'default'"
            :ghost="currentPage !== 'traces'"
            @click="currentPage = 'traces'; loadAllTraces()"
          >
            <template #icon>
              <n-icon><ChatboxEllipsesIcon /></n-icon>
            </template>
            {{ t('nav.traces') }}
          </n-button>
        </div>

        <div style="display: flex; align-items: center; gap: 16px;">
          <n-button quaternary circle size="small" @click="refreshAll" :loading="refreshing">
            <template #icon>
              <n-icon :size="20">
                <RefreshIcon />
              </n-icon>
            </template>
          </n-button>

          <n-button quaternary circle size="small" @click="currentPage = 'settings'">
            <template #icon>
              <n-icon :size="20">
                <SettingsIcon />
              </n-icon>
            </template>
          </n-button>

          <n-button quaternary circle size="small" @click="toggleTheme">
            <template #icon>
              <n-icon>
                <MoonIcon v-if="isDark" />
                <SunnyIcon v-else />
              </n-icon>
            </template>
          </n-button>

          <n-button quaternary circle size="small" @click="showLanguageModal = true">
            <template #icon>
              <n-icon :size="20">
                <LanguageIcon />
              </n-icon>
            </template>
          </n-button>

          <n-button type="primary" size="small" @click="showAddModal = true">
            <template #icon>
              <n-icon><AddIcon /></n-icon>
            </template>
            {{ t('nav.addRoute') }}
          </n-button>
        </div>
      </n-layout-header>

      <!-- Main Content -->
      <n-layout-content style="padding: 24px; overflow: auto;">
        <!-- Home Page -->
        <div v-if="currentPage === 'home'">
          <!-- Stats Cards -->
          <n-grid :cols="4" :x-gap="16" :y-gap="16" style="margin-bottom: 24px;">
            <n-grid-item>
              <n-card :bordered="false" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
                <n-statistic :label="t('home.routeCount')" :value="stats.route_count">
                  <template #prefix>
                    <n-icon size="24" color="#fff">
                      <GitNetworkIcon />
                    </n-icon>
                  </template>
                </n-statistic>
              </n-card>
            </n-grid-item>

            <n-grid-item>
              <n-card :bordered="false" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
                <n-statistic :label="t('home.modelCount')" :value="stats.model_count">
                  <template #prefix>
                    <n-icon size="24" color="#fff">
                      <CubeIcon />
                    </n-icon>
                  </template>
                </n-statistic>
              </n-card>
            </n-grid-item>

            <n-grid-item>
              <n-card :bordered="false" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
                <n-statistic :label="t('home.todayRequests')" :value="stats.today_requests || 0">
                  <template #prefix>
                    <n-icon size="24" color="#fff">
                      <StatsChartIcon />
                    </n-icon>
                  </template>
                </n-statistic>
              </n-card>
            </n-grid-item>

            <n-grid-item>
              <n-card :bordered="false" style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);">
                <n-statistic :label="t('home.todayTokens')" :value="formatNumber(stats.today_tokens || 0)">
                  <template #prefix>
                    <n-icon size="24" color="#fff">
                      <FlashIcon />
                    </n-icon>
                  </template>
                </n-statistic>
              </n-card>
            </n-grid-item>
          </n-grid>

          <!-- Redirect Config -->
          <n-card :title="'🔀 ' + t('home.redirectConfig')" style="margin-bottom: 24px;" :bordered="false">
            <n-space vertical>
              <n-space align="center">
                <span>{{ t('home.enableRedirect') }}:</span>
                <n-switch v-model:value="redirectConfig.enabled" @update:value="saveRedirectConfig" />
              </n-space>

              <n-space align="center" v-if="redirectConfig.enabled">
                <n-tag type="info" size="large" style="font-family: monospace;">
                  {{ redirectConfig.keyword }}
                </n-tag>
                <n-icon size="20"><ArrowForwardIcon /></n-icon>
                <n-tag type="success" size="large" style="font-family: monospace;">
                  {{ redirectConfig.targetModel || t('home.notConfigured') }}
                </n-tag>
                <n-tag v-if="redirectConfig.targetName" type="warning" size="large">
                  ({{ redirectConfig.targetName }})
                </n-tag>
                <!-- 跳转按钮 -->
                <n-button
                  v-if="redirectConfig.targetModel"
                  size="small"
                  @click="jumpToTargetModel"
                >
                  <template #icon>
                    <n-icon><LocationIcon /></n-icon>
                  </template>
                  {{ t('home.jumpToTarget') }}
                </n-button>
              </n-space>
            </n-space>
          </n-card>

          <!-- API Config -->
          <n-card :title="'🔑 ' + t('home.apiConfig')" style="margin-bottom: 24px;" :bordered="false">
            <n-grid :cols="2" :x-gap="24">
              <!-- 左侧: OpenAI 兼容接口 -->
              <n-grid-item>
                <n-space vertical :size="12">
                  <n-text strong style="font-size: 14px;">{{ t('home.openaiInterface') }}</n-text>
                  <n-text depth="3" style="font-size: 12px;">{{ t('home.openaiInterfaceDesc') }}</n-text>

                  <div>
                    <n-text depth="2" style="font-size: 13px; margin-bottom: 4px; display: block;">{{ t('home.apiAddress') }}</n-text>
                    <n-input
                      :value="config.localApiEndpoint + '/api'"
                      readonly
                      size="small"
                    >
                      <template #suffix>
                        <n-button text size="small" @click="copyToClipboard(config.localApiEndpoint + '/api')">
                          <template #icon>
                            <n-icon><CopyIcon /></n-icon>
                          </template>
                        </n-button>
                      </template>
                    </n-input>
                    <n-text depth="3" style="font-size: 11px; margin-top: 4px; display: block; color: #18a058;">
                      📝 {{ t('home.openaiPath') }}：{{ config.localApiEndpoint }}/api/v1/chat/completions
                    </n-text>
                  </div>

                  <div>
                    <n-text depth="2" style="font-size: 13px; margin-bottom: 4px; display: block;">{{ t('home.apiKey') }}</n-text>
                    <n-input
                      :value="maskApiKey(config.localApiKey)"
                      readonly
                      size="small"
                    >
                      <template #suffix>
                        <n-button text size="small" @click="copyToClipboard(config.localApiKey)" :title="t('home.copyApiKey')">
                          <template #icon>
                            <n-icon><CopyIcon /></n-icon>
                          </template>
                        </n-button>
                        <n-button text size="small" @click="showEditApiKeyModal = true" style="margin-left: 8px;" :title="t('home.editApiKey')">
                          <template #icon>
                            <n-icon><EditIcon /></n-icon>
                          </template>
                        </n-button>
                        <n-button text size="small" @click="generateNewApiKey" style="margin-left: 8px;" :title="t('home.randomApiKey')">
                          <template #icon>
                            <n-icon><RefreshIcon /></n-icon>
                          </template>
                        </n-button>
                      </template>
                    </n-input>
                  </div>
                </n-space>
              </n-grid-item>

              <!-- 右侧: 翻译 API 接口 -->
              <n-grid-item>
                <n-space vertical :size="12">
                  <n-text strong style="font-size: 14px;">{{ t('home.translationInterface') }}</n-text>
                  <n-text depth="3" style="font-size: 12px;">{{ t('home.translationInterfaceDesc') }}</n-text>

                  <div>
                    <n-text depth="2" style="font-size: 13px; margin-bottom: 4px; display: block;">{{ t('home.claudeCodeInterface') }}</n-text>
                    <n-input
                      :value="config.localApiEndpoint + '/api/claudecode'"
                      readonly
                      size="small"
                    >
                      <template #suffix>
                        <n-button text size="small" @click="copyToClipboard(config.localApiEndpoint + '/api/claudecode')">
                          <template #icon>
                            <n-icon><CopyIcon /></n-icon>
                          </template>
                        </n-button>
                      </template>
                    </n-input>
                    <n-text depth="3" style="font-size: 11px; margin-top: 4px; display: block; color: #18a058;">
                      📝 {{ t('home.claudeCodePath') }}：{{ config.localApiEndpoint }}/api/claudecode/v1/messages
                    </n-text>
                  </div>

                  <div>
                    <n-text depth="2" style="font-size: 13px; margin-bottom: 4px; display: block;">{{ t('home.anthropicInterface') }}</n-text>
                    <n-input
                      :value="config.localApiEndpoint + '/api/anthropic'"
                      readonly
                      size="small"
                    >
                      <template #suffix>
                        <n-button text size="small" @click="copyToClipboard(config.localApiEndpoint + '/api/anthropic')">
                          <template #icon>
                            <n-icon><CopyIcon /></n-icon>
                          </template>
                        </n-button>
                      </template>
                    </n-input>
                    <n-text depth="3" style="font-size: 11px; margin-top: 4px; display: block; color: #18a058;">
                      📝 {{ t('home.anthropicPath') }}：{{ config.localApiEndpoint }}/api/anthropic/v1/messages
                    </n-text>
                  </div>

                  <div>
                    <n-text depth="2" style="font-size: 13px; margin-bottom: 4px; display: block;">{{ t('home.geminiInterface') }}</n-text>
                    <n-input
                      :value="config.localApiEndpoint + '/api/gemini'"
                      readonly
                      size="small"
                    >
                      <template #suffix>
                        <n-button text size="small" @click="copyToClipboard(config.localApiEndpoint + '/api/gemini')">
                          <template #icon>
                            <n-icon><CopyIcon /></n-icon>
                          </template>
                        </n-button>
                      </template>
                    </n-input>
                    <n-text depth="3" style="font-size: 11px; margin-top: 4px; display: block; color: #18a058;">
                      📝 {{ t('home.geminiPath') }}：{{ config.localApiEndpoint }}/api/gemini/completions
                    </n-text>
                  </div>

                  <div>
                    <n-text depth="2" style="font-size: 13px; margin-bottom: 4px; display: block;">{{ t('home.cursorInterface') }}</n-text>
                    <n-input
                      :value="config.localApiEndpoint + '/api/cursor/v1'"
                      readonly
                      size="small"
                    >
                      <template #suffix>
                        <n-button text size="small" @click="copyToClipboard(config.localApiEndpoint + '/api/cursor/v1')">
                          <template #icon>
                            <n-icon><CopyIcon /></n-icon>
                          </template>
                        </n-button>
                      </template>
                    </n-input>
                    <n-text depth="3" style="font-size: 11px; margin-top: 4px; display: block; color: #18a058;">
                      📝 {{ t('home.cursorPath') }}：{{ config.localApiEndpoint }}/api/cursor/v1/chat/completions
                    </n-text>
                  </div>
                </n-space>
              </n-grid-item>
            </n-grid>
          </n-card>
        </div>

        <!-- Models Page -->
        <div v-if="currentPage === 'models'">
          <n-card :title="'📋 ' + t('models.title')" :bordered="false">
            <template #header-extra>
              <n-space>
                <n-button @click="exportRoutes" type="primary" ghost size="small">
                  <template #icon>
                    <n-icon><ArrowForwardIcon style="transform: rotate(-90deg);" /></n-icon>
                  </template>
                  {{ t('models.exportJson') }}
                </n-button>
                <n-button @click="triggerImport" type="primary" ghost size="small">
                  <template #icon>
                    <n-icon><ArrowForwardIcon style="transform: rotate(90deg);" /></n-icon>
                  </template>
                  {{ t('models.importJson') }}
                </n-button>
                <n-button @click="loadRoutes" quaternary circle size="small">
                  <template #icon>
                    <n-icon><RefreshIcon /></n-icon>
                  </template>
                </n-button>
              </n-space>
              <input
                ref="fileInput"
                type="file"
                accept=".json"
                style="display: none;"
                @change="handleFileImport"
              />
            </template>

            <!-- 按分组显示的折叠面板 -->
            <n-collapse v-model:expanded-names="expandedGroups">
              <n-collapse-item
                v-for="(groupRoutes, groupName) in groupedRoutes"
                :key="groupName"
                :name="groupName"
                :title="`${t('models.group')}: ${groupName || t('models.ungrouped')} (${groupRoutes.length} ${t('models.modelCount')})`"
              >
                <n-data-table
                  :columns="modelsPageColumns"
                  :data="groupRoutes"
                  :bordered="false"
                  :single-line="false"
                  size="small"
                  striped
                  :pagination="false"
                  :row-props="rowProps"
                />
              </n-collapse-item>
            </n-collapse>

            <n-empty
              v-if="routes.length === 0"
              :description="t('models.noRoutes')"
              style="margin: 60px 0;"
            />
          </n-card>
        </div>

        <!-- Stats Page -->
        <div v-if="currentPage === 'stats'">
          <n-space vertical :size="16">
            <!-- 今日消耗统计卡片 -->
            <n-card :title="'📊 ' + t('stats.todayStats')" :bordered="false">
              <template #header-extra>
                <n-space>
                  <n-tooltip trigger="hover">
                    <template #trigger>
                      <n-button type="warning" quaternary size="small" @click="compressDatabase" :loading="compressing" :disabled="compressing">
                        <template #icon>
                          <n-icon><ArchiveIcon /></n-icon>
                        </template>
                        {{ compressing ? t('settings.compressing') : t('settings.compressDatabase') }}
                      </n-button>
                    </template>
                    {{ t('settings.compressDatabaseDesc') }}
                  </n-tooltip>
                  <n-button type="error" quaternary size="small" @click="showClearStatsDialog">
                    <template #icon>
                      <n-icon><TrashIcon /></n-icon>
                    </template>
                    {{ t('stats.clearData') }}
                  </n-button>
                </n-space>
              </template>
              <n-grid :cols="4" :x-gap="16">
                <n-grid-item>
                  <n-statistic :label="t('stats.todayTokens')" :value="formatNumber(stats.today_tokens || 0)">
                    <template #prefix>
                      <n-icon size="20" color="#18a058">
                        <FlashIcon />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-grid-item>
                <n-grid-item>
                  <n-statistic :label="t('stats.todayRequests')" :value="stats.today_requests || 0">
                    <template #prefix>
                      <n-icon size="20" color="#18a058">
                        <StatsChartIcon />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-grid-item>
                <n-grid-item>
                  <n-statistic :label="t('stats.totalTokens')" :value="formatNumber(stats.total_tokens)">
                    <template #prefix>
                      <n-icon size="20" color="#18a058">
                        <FlashIcon />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-grid-item>
                <n-grid-item>
                  <n-statistic :label="t('stats.totalRequests')" :value="stats.total_requests">
                    <template #prefix>
                      <n-icon size="20" color="#18a058">
                        <StatsChartIcon />
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-grid-item>
              </n-grid>
            </n-card>

            <!-- GitHub 热力图样式的历史使用量 -->
            <n-card :title="'🔥 ' + t('stats.heatmap')" :bordered="false">
              <div class="heatmap-container" @mouseleave="heatmapTooltip.show = false">
                <div class="heatmap-months-row">
                  <span 
                    v-for="monthData in heatmapMonthsWithPosition" 
                    :key="monthData.weekIndex"
                    class="heatmap-month-label"
                    :style="{ left: (monthData.weekIndex / 53 * 100) + '%' }"
                  >{{ monthData.name }}</span>
                </div>
                <div class="heatmap-grid">
                  <div v-for="(week, weekIndex) in heatmapData" :key="weekIndex" class="heatmap-week">
                    <div
                      v-for="(day, dayIndex) in week"
                      :key="dayIndex"
                      class="heatmap-cell"
                      :class="getHeatmapClass(day.tokens)"
                      @mouseenter="showHeatmapTooltip($event, day)"
                      @mouseleave="heatmapTooltip.show = false"
                    ></div>
                  </div>
                </div>
                <!-- 单一 tooltip 元素 -->
                <div 
                  v-show="heatmapTooltip.show" 
                  class="heatmap-tooltip"
                  :style="{ left: heatmapTooltip.x + 'px', top: heatmapTooltip.y + 'px' }"
                >
                  <div style="font-weight: bold;">{{ t('stats.date') }}: {{ heatmapTooltip.date }}</div>
                  <div>{{ t('stats.inputTokens') }}: {{ formatNumber(heatmapTooltip.requestTokens) }}</div>
                  <div>{{ t('stats.outputTokens') }}: {{ formatNumber(heatmapTooltip.responseTokens) }}</div>
                  <div>{{ t('stats.totalTokensCol') }}: {{ formatNumber(heatmapTooltip.tokens) }}</div>
                  <div>{{ t('stats.requestCount') }}: {{ heatmapTooltip.requests }}</div>
                </div>
                <div class="heatmap-legend">
                  <span>{{ t('stats.less') }}</span>
                  <div class="legend-box level-0"></div>
                  <div class="legend-box level-1"></div>
                  <div class="legend-box level-2"></div>
                  <div class="legend-box level-3"></div>
                  <div class="legend-box level-4"></div>
                  <span>{{ t('stats.more') }}</span>
                </div>
              </div>
            </n-card>

            <!-- 今日按时间段显示的折线图 -->
            <n-card :title="'📈 ' + t('stats.todayTrend')" :bordered="false">
              <v-chart :option="todayChartOption" style="height: 300px;" :theme="isDark ? 'dark' : ''" autoresize />
            </n-card>

            <!-- 历史使用量 - 接口使用排行 -->
            <n-card :title="'🏆 ' + t('stats.modelRanking')" :bordered="false">
              <n-data-table
                :columns="rankingColumns"
                :data="modelRankingData"
                :pagination="false"
                :bordered="false"
                striped
                size="small"
              />
            </n-card>

            <!-- 用量汇总统计 -->
            <n-card :title="'📊 ' + t('stats.usageSummary')" :bordered="false">
              <n-grid :cols="3" :x-gap="16" :y-gap="16">
                <!-- 周用量 -->
                <n-grid-item>
                  <n-card :title="t('stats.weeklyUsage')" size="small" :bordered="true">
                    <n-data-table
                      :columns="weeklyColumns"
                      :data="usageSummary.week_stats || []"
                      :pagination="false"
                      :bordered="false"
                      size="small"
                      :max-height="200"
                    />
                  </n-card>
                </n-grid-item>

                <!-- 年用量 -->
                <n-grid-item>
                  <n-card :title="t('stats.yearlyUsage')" size="small" :bordered="true">
                    <n-data-table
                      :columns="yearlyColumns"
                      :data="usageSummary.year_stats || []"
                      :pagination="false"
                      :bordered="false"
                      size="small"
                      :max-height="200"
                    />
                  </n-card>
                </n-grid-item>

                <!-- 总用量 -->
                <n-grid-item>
                  <n-card :title="t('stats.totalUsage')" size="small" :bordered="true">
                    <n-space vertical :size="8">
                      <n-statistic :label="t('stats.totalRequests')" :value="usageSummary.total_stats?.request_count || 0" />
                      <n-statistic :label="t('stats.totalTokensCol')" :value="formatNumber(usageSummary.total_stats?.total_tokens || 0)" />
                      <n-statistic :label="t('stats.inputTokens')" :value="formatNumber(usageSummary.total_stats?.request_tokens || 0)" />
                      <n-statistic :label="t('stats.outputTokens')" :value="formatNumber(usageSummary.total_stats?.response_tokens || 0)" />
                      <n-space>
                        <n-tag type="success" size="small">{{ t('stats.successCount') }}: {{ usageSummary.total_stats?.success_count || 0 }}</n-tag>
                        <n-tag type="error" size="small">{{ t('stats.failCount') }}: {{ usageSummary.total_stats?.fail_count || 0 }}</n-tag>
                      </n-space>
                    </n-space>
                  </n-card>
                </n-grid-item>
              </n-grid>
            </n-card>
          </n-space>
        </div>

        <!-- Logs Page -->
        <div v-if="currentPage === 'logs'">
          <n-card :title="'📋 ' + t('logs.title')" :bordered="false">
            <template #header-extra>
              <n-space>
                <n-button quaternary circle size="small" @click="loadRequestLogs" :loading="logsLoading">
                  <template #icon>
                    <n-icon><RefreshIcon /></n-icon>
                  </template>
                </n-button>
              </n-space>
            </template>

            <!-- 筛选器 -->
            <n-space style="margin-bottom: 16px;" align="center">
              <n-input
                v-model:value="logsFilter.model"
                :placeholder="t('logs.filterModel')"
                style="width: 180px;"
                size="small"
                clearable
                @update:value="debounceLoadLogs"
              >
                <template #prefix>
                  <n-icon><SearchIcon /></n-icon>
                </template>
              </n-input>
              <n-select
                v-model:value="logsFilter.style"
                :placeholder="t('logs.filterStyle')"
                :options="styleOptions"
                style="width: 140px;"
                size="small"
                clearable
                @update:value="loadRequestLogs"
              />
              <n-select
                v-model:value="logsFilter.success"
                :placeholder="t('logs.filterStatus')"
                :options="successOptions"
                style="width: 120px;"
                size="small"
                clearable
                @update:value="loadRequestLogs"
              />
              <n-button size="small" @click="clearLogsFilter">
                {{ t('logs.clearFilter') }}
              </n-button>
            </n-space>

            <!-- 日志表格 -->
            <n-data-table
              :columns="logsColumns"
              :data="logsData"
              :bordered="false"
              :single-line="false"
              size="small"
              striped
              :loading="logsLoading"
              :pagination="false"
              :max-height="500"
            />

            <!-- 分页 -->
            <n-space justify="center" style="margin-top: 16px;">
              <n-pagination
                v-model:page="logsPage"
                :page-size="logsPageSize"
                :item-count="logsTotal"
                :page-sizes="[20, 50, 100]"
                show-size-picker
                @update:page="loadRequestLogs"
                @update:page-size="handleLogsPageSizeChange"
              />
            </n-space>

            <n-empty
              v-if="logsData.length === 0 && !logsLoading"
              :description="t('logs.noLogs')"
              style="margin: 40px 0;"
            />
          </n-card>
        </div>

        <!-- Health Page -->
        <div v-if="currentPage === 'health'">
          <n-card :title="'📊 ' + t('health.title')" :bordered="false">
            <template #header-extra>
              <n-button quaternary circle size="small" @click="loadHealthStatus" :loading="healthLoading">
                <template #icon>
                  <n-icon><RefreshIcon /></n-icon>
                </template>
              </n-button>
            </template>

            <n-spin :show="healthLoading">
              <n-space vertical :size="24">
                <!-- 按分组显示 -->
                <div v-for="group in healthData" :key="group.group">
                  <n-card size="small" :bordered="true">
                    <template #header>
                      <n-space align="center">
                        <n-text strong style="font-size: 16px;">
                          {{ group.group === 'default' ? t('models.ungrouped') : group.group }}
                        </n-text>
                        <n-tag type="info" size="small">{{ group.route_count }} {{ t('health.routes') }}</n-tag>
                        <n-tag :type="group.success_rate >= 90 ? 'success' : group.success_rate >= 70 ? 'warning' : 'error'" size="small">
                          {{ t('health.successRate') }}: {{ group.success_rate.toFixed(1) }}%
                        </n-tag>
                      </n-space>
                    </template>

                    <!-- 路由列表 -->
                    <n-space vertical :size="12">
                      <div v-for="route in group.routes" :key="route.id" class="health-route-item">
                        <n-space align="center" justify="space-between" style="width: 100%;">
                          <n-space align="center" style="min-width: 200px;">
                            <n-text strong>{{ route.name }}</n-text>
                            <n-text depth="3" style="font-size: 12px;">{{ route.model }}</n-text>
                          </n-space>
                          
                          <!-- 状态条 -->
                          <div class="status-bar-container">
                            <div class="status-bar" v-if="route.status_history && route.status_history.length > 0">
                              <span
                                v-for="(success, idx) in route.status_history"
                                :key="idx"
                                class="status-dot"
                                :class="success ? 'success' : 'fail'"
                                :title="success ? t('logs.success') : t('logs.failed')"
                              ></span>
                            </div>
                            <n-text v-else depth="3" style="font-size: 12px;">{{ t('health.noData') }}</n-text>
                          </div>

                          <n-space align="center" style="min-width: 150px;">
                            <n-tag :type="route.success_rate >= 90 ? 'success' : route.success_rate >= 70 ? 'warning' : 'error'" size="small">
                              {{ route.success_rate.toFixed(1) }}%
                            </n-tag>
                            <n-text depth="3" style="font-size: 12px;">{{ route.total_requests }} {{ t('health.totalRequests') }}</n-text>
                          </n-space>
                        </n-space>
                      </div>
                    </n-space>
                  </n-card>
                </div>

                <n-empty
                  v-if="healthData.length === 0 && !healthLoading"
                  :description="t('health.noData')"
                  style="margin: 40px 0;"
                />
              </n-space>
            </n-spin>
          </n-card>
        </div>

        <!-- Traces Page -->
        <div v-if="currentPage === 'traces'">
          <n-card :title="'💬 ' + t('traces.title')" :bordered="false">
            <template #header-extra>
              <n-space align="center">
                <n-input
                  v-model:value="tracesSearchQuery"
                  :placeholder="t('traces.search')"
                  clearable
                  size="small"
                  style="width: 200px;"
                  @update:value="debounceSearchTraces"
                >
                  <template #prefix>
                    <n-icon><SearchIcon /></n-icon>
                  </template>
                </n-input>
                <n-checkbox v-model:checked="tracesAutoRefresh" size="small" @update:checked="toggleTracesAutoRefresh">
                  {{ t('traces.autoRefresh') }}
                </n-checkbox>
                <n-button quaternary circle size="small" @click="loadAllTraces" :loading="tracesLoading">
                  <template #icon>
                    <n-icon><RefreshIcon /></n-icon>
                  </template>
                </n-button>
                <n-popconfirm @positive-click="clearAllTraces">
                  <template #trigger>
                    <n-button quaternary circle size="small" type="error">
                      <template #icon>
                        <n-icon><TrashIcon /></n-icon>
                      </template>
                    </n-button>
                  </template>
                  {{ t('traces.confirmClearAll') }}
                </n-popconfirm>
              </n-space>
            </template>

            <n-spin :show="tracesLoading">
              <n-list bordered style="max-height: calc(100vh - 320px); overflow-y: auto;">
                <n-list-item
                  v-for="trace in filteredTraces"
                  :key="trace.id"
                  style="padding: 12px 16px;"
                >
                  <n-space vertical :size="8" style="width: 100%;">
                    <!-- 头部信息 -->
                    <n-space align="center" justify="space-between" style="width: 100%;">
                      <n-space align="center" size="small">
                        <n-tag :type="trace.success ? 'success' : 'error'" size="small" round>
                          {{ trace.success ? '✓' : '✗' }}
                        </n-tag>
                        <n-text strong style="font-size: 14px;">{{ trace.created_at }}</n-text>
                        <n-text depth="3" style="font-size: 12px;">{{ trace.remote_ip }}</n-text>
                        <n-text depth="3" style="font-size: 12px;">{{ trace.proxy_time_ms }}ms</n-text>
                      </n-space>
                      <n-space align="center" size="small">
                        <n-text strong>{{ trace.model }}</n-text>
                        <n-tag v-if="trace.is_stream" size="tiny" type="warning">流式</n-tag>
                        <n-text depth="3" style="font-size: 12px;">{{ trace.provider_name }}</n-text>
                      </n-space>
                    </n-space>

                    <!-- 详情展开 -->
                    <n-collapse arrow-placement="left" :default-expanded-names="[]">
                      <n-collapse-item :title="t('traces.request')" name="request">
                        <n-code :code="formatJson(trace.request_content)" language="json" word-wrap style="max-height: 300px; overflow-y: auto;" />
                      </n-collapse-item>
                      <n-collapse-item :title="t('traces.response')" name="response">
                        <n-code :code="formatJson(trace.response_content)" language="json" word-wrap style="max-height: 300px; overflow-y: auto;" />
                      </n-collapse-item>
                    </n-collapse>

                    <!-- token 信息 -->
                    <n-space v-if="trace.total_tokens > 0" size="small">
                      <n-text depth="3" style="font-size: 11px;">tokens: {{ trace.request_tokens }}/{{ trace.response_tokens }}/{{ trace.total_tokens }}</n-text>
                    </n-space>
                    <n-text v-if="trace.error_message && !trace.success" type="error" style="font-size: 12px;">
                      {{ trace.error_message.slice(0, 200) }}{{ trace.error_message.length > 200 ? '...' : '' }}
                    </n-text>
                  </n-space>
                </n-list-item>
                <n-empty v-if="allTraces.length === 0 && !tracesLoading" :description="t('traces.noTraces')" style="padding: 40px;" />
              </n-list>
            </n-spin>

            <!-- 分页 -->
            <n-space justify="center" style="margin-top: 16px;" v-if="allTracesTotal > 0">
              <n-pagination
                v-model:page="allTracesPage"
                :page-size="allTracesPageSize"
                :item-count="allTracesTotal"
                show-size-picker
                :page-sizes="[20, 50, 100]"
                @update:page="loadAllTraces"
                @update:page-size="handleTracesPageSizeChange"
              />
            </n-space>
          </n-card>
        </div>

        <!-- Settings Page -->
        <div v-if="currentPage === 'settings'">
          <n-card :title="'⚙️ ' + t('settings.title')" :bordered="false">
            <n-space vertical :size="24">
              <!-- GitHub 项目信息 -->
              <div>
                <n-text strong style="font-size: 16px;">{{ t('settings.projectInfo') }}</n-text>
                <n-space vertical :size="12" style="margin-top: 12px;">
                  <n-space align="center">
                    <n-icon size="20"><LogoGithubIcon /></n-icon>
                    <n-text>{{ t('settings.githubRepo') }}:</n-text>
                    <n-button text type="primary" size="small" tag="a" href="https://github.com/cniu6/anyproxyai" target="_blank">
                      github.com/cniu6/anyproxyai
                    </n-button>
                  </n-space>

                  <n-space align="center">
                    <n-icon size="20"><InformationCircleIcon /></n-icon>
                    <n-text>{{ t('settings.version') }}: v2.0.7</n-text>
                  </n-space>

                  <n-space align="center">
                    <n-icon size="20"><CodeIcon /></n-icon>
                    <n-text>{{ t('settings.builtWith') }}</n-text>
                  </n-space>
                </n-space>
              </div>

              <n-divider />

              <!-- 应用选项 -->
              <div>
                <n-text strong style="font-size: 16px;">{{ t('settings.appOptions') }}</n-text>
                <n-space vertical :size="16" style="margin-top: 12px;">
                  <!-- 重定向关键字设置 -->
                  <div>
                    <n-text depth="2" style="font-size: 14px; margin-bottom: 8px; display: block;">{{ t('settings.redirectKeyword') }}</n-text>
                    <n-input
                      v-model:value="settings.redirectKeyword"
                      placeholder="proxy_auto"
                      style="max-width: 300px;"
                    >
                      <template #suffix>
                        <n-button text size="small" @click="updateRedirectKeyword">
                          {{ t('settings.save') }}
                        </n-button>
                      </template>
                    </n-input>
                    <n-text depth="3" style="font-size: 12px; margin-top: 4px; display: block;">
                      {{ t('settings.redirectKeywordDesc') }}
                    </n-text>
                  </div>

                  <n-checkbox v-model:checked="settings.autoStart" @update:checked="toggleAutoStart">
                    {{ t('settings.autoStart') }}
                  </n-checkbox>

                  <n-checkbox v-model:checked="settings.minimizeToTray" @update:checked="toggleMinimizeToTray">
                    {{ t('settings.minimizeToTray') }}
                  </n-checkbox>

                  <n-checkbox v-model:checked="settings.enableFileLog" @update:checked="toggleEnableFileLog">
                    {{ t('settings.enableFileLog') }}
                  </n-checkbox>
                  <n-text depth="3" style="font-size: 12px; margin-left: 24px;">
                    {{ t('settings.enableFileLogDesc') }}
                  </n-text>

                  <n-checkbox v-model:checked="settings.fallbackEnabled" @update:checked="toggleFallbackEnabled">
                    {{ t('settings.enableFallback') }}
                  </n-checkbox>
                  <n-text depth="3" style="font-size: 12px; margin-left: 24px;">
                    {{ t('settings.enableFallbackDesc') }}
                  </n-text>

                  <n-checkbox v-model:checked="settings.tracesEnabled" @update:checked="toggleTracesEnabled">
                    {{ t('settings.enableTraces') }}
                  </n-checkbox>
                  <n-text depth="3" style="font-size: 12px; margin-left: 24px;">
                    {{ t('settings.enableTracesDesc') }}
                  </n-text>

                  <!-- Traces 保留天数 -->
                  <div v-if="settings.tracesEnabled" style="margin-left: 24px; margin-top: 8px;">
                    <n-space align="center">
                      <n-text depth="2" style="font-size: 13px;">{{ t('settings.tracesRetentionDays') }}:</n-text>
                      <n-input-number
                        v-model:value="settings.tracesRetentionDays"
                        :min="1"
                        :max="365"
                        size="small"
                        style="width: 100px;"
                        @blur="updateTracesRetentionDays"
                      />
                      <n-text depth="3" style="font-size: 12px;">{{ t('settings.days') }}</n-text>
                    </n-space>
                  </div>

                  <!-- API 端口设置 -->
                  <div style="margin-top: 16px;">
                    <n-text depth="2" style="font-size: 14px; margin-bottom: 8px; display: block;">{{ t('settings.apiPort') }}</n-text>
                    <n-input-number
                      v-model:value="settings.port"
                      :min="1"
                      :max="65535"
                      style="max-width: 200px;"
                    >
                      <template #suffix>
                        <n-button text size="small" @click="updatePort">
                          {{ t('settings.save') }}
                        </n-button>
                      </template>
                    </n-input-number>
                    <n-text depth="3" style="font-size: 12px; margin-top: 4px; display: block;">
                      {{ t('settings.apiPortDesc') }}
                    </n-text>
                  </div>
                </n-space>
              </div>

              <n-divider />

              <!-- 语言设置 -->
              <div>
                <n-text strong style="font-size: 16px;">{{ t('settings.languageSettings') }}</n-text>
                <n-space align="center" style="margin-top: 12px;">
                  <n-text>{{ t('settings.language') }}:</n-text>
                  <n-select
                    :value="currentLocale"
                    @update:value="switchLanguage"
                    :options="[
                      { label: '🇨🇳 简体中文', value: 'zh-CN' },
                      { label: '🇺🇸 English', value: 'en-US' }
                    ]"
                    style="width: 160px;"
                  />
                </n-space>
                <n-text depth="3" style="font-size: 12px; margin-top: 4px; display: block;">
                  {{ t('settings.languageDesc') }}
                </n-text>
              </div>

              <n-divider />

              <!-- 主题设置 -->
              <div>
                <n-text strong style="font-size: 16px;">{{ t('settings.themeSettings') }}</n-text>
                <n-space align="center" style="margin-top: 12px;">
                  <n-text>{{ t('settings.currentTheme') }}:</n-text>
                  <n-tag :type="isDark ? 'info' : 'warning'" size="small">
                    {{ isDark ? t('settings.darkMode') : t('settings.lightMode') }}
                  </n-tag>
                  <n-button size="small" @click="toggleTheme">
                    <template #icon>
                      <n-icon>
                        <MoonIcon v-if="!isDark" />
                        <SunnyIcon v-else />
                      </n-icon>
                    </template>
                    {{ t('settings.switchTheme') }}
                  </n-button>
                </n-space>
              </div>
            </n-space>
          </n-card>
        </div>
      </n-layout-content>
    </n-layout>

    <!-- Add Route Modal -->
    <AddRouteModal 
      v-model:visible="showAddModal" 
      @route-added="handleRouteAdded" 
    />
    
    <!-- Edit Route Modal -->
    <EditRouteModal
      v-model:visible="showEditModal"
      :route="editingRoute"
      @route-updated="handleRouteUpdated"
    />

    <!-- Language Switch Modal -->
    <n-modal
      v-model:show="showLanguageModal"
      preset="card"
      :title="t('settings.language')"
      style="width: 400px;"
      :bordered="false"
    >
      <n-space vertical :size="16">
        <n-text depth="3">{{ t('settings.languageDesc') }}</n-text>
        <n-radio-group :value="currentLocale" @update:value="switchLanguage">
          <n-space vertical>
            <n-radio value="zh-CN" size="large">
              🇨🇳 简体中文
            </n-radio>
            <n-radio value="en-US" size="large">
              🇺🇸 English
            </n-radio>
          </n-space>
        </n-radio-group>
      </n-space>
    </n-modal>

    <!-- Clear Stats Confirmation Dialog -->
    <n-modal
      v-model:show="showClearDialog"
      preset="dialog"
      :title="t('clearDialog.title')"
      type="error"
      :positive-text="t('clearDialog.confirm')"
      :negative-text="t('clearDialog.cancel')"
      @positive-click="confirmClearStats"
      @negative-click="showClearDialog = false"
    >
      <template #icon>
        <n-icon size="24" color="#e88080">
          <TrashIcon />
        </n-icon>
      </template>
      {{ t('clearDialog.message') }}
      <br>
      <br>
      <strong>{{ t('clearDialog.dataInclude') }}</strong>
      <ul>
        <li>{{ t('clearDialog.requestLogs') }}</li>
        <li>{{ t('clearDialog.tokenStats') }}</li>
        <li>{{ t('clearDialog.modelRanking') }}</li>
        <li>{{ t('clearDialog.heatmapData') }}</li>
      </ul>
    </n-modal>

    <!-- Restart Confirmation Dialog -->
    <n-modal
      v-model:show="showRestartDialog"
      preset="dialog"
      :title="t('restartDialog.title')"
      type="warning"
      :positive-text="t('restartDialog.confirm')"
      :negative-text="t('restartDialog.cancel')"
      @positive-click="restartApp"
      @negative-click="showRestartDialog = false"
    >
      <template #icon>
        <n-icon size="24" color="#f0a020">
          <RefreshIcon />
        </n-icon>
      </template>
      {{ t('restartDialog.message') }}
    </n-modal>

    <!-- Edit API Key Dialog -->
    <n-modal
      v-model:show="showEditApiKeyModal"
      preset="card"
      :title="t('home.editApiKeyTitle')"
      style="width: 500px;"
      :bordered="false"
    >
      <n-space vertical :size="16">
        <n-text depth="3">{{ t('home.editApiKeyDesc') }}</n-text>
        <n-input
          v-model:value="newApiKey"
          type="text"
          :placeholder="t('home.apiKeyPlaceholder')"
          clearable
        />
        <n-space justify="end">
          <n-button @click="showEditApiKeyModal = false">{{ t('addRoute.cancel') }}</n-button>
          <n-button type="primary" @click="saveCustomApiKey" :disabled="!newApiKey.trim()">
            {{ t('settings.save') }}
          </n-button>
        </n-space>
      </n-space>
    </n-modal>
  </n-config-provider>
</template>

<script setup>
import { ref, h, onMounted, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { darkTheme, NButton, NIcon, NTag, NSpace, NModal, NTooltip, NSwitch } from 'naive-ui'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
} from 'echarts/components'
import {
  ServerOutline as ServerIcon,
  Moon as MoonIcon,
  Sunny as SunnyIcon,
  Add as AddIcon,
  GitNetwork as GitNetworkIcon,
  Cube as CubeIcon,
  StatsChart as StatsChartIcon,
  Flash as FlashIcon,
  ArrowForward as ArrowForwardIcon,
  Copy as CopyIcon,
  Refresh as RefreshIcon,
  CreateOutline as EditIcon,
  TrashOutline as DeleteIcon,
  Home as HomeIcon,
  List as ListIcon,
  BarChart as BarChartIcon,
  Settings as SettingsIcon,
  Location as LocationIcon,
  LogoGithub as LogoGithubIcon,
  InformationCircle as InformationCircleIcon,
  Code as CodeIcon,
  Link as LinkIcon,
  Trash as TrashIcon,
  Language as LanguageIcon,
  Archive as ArchiveIcon,
  DocumentText as DocumentTextIcon,
  TimeOutline as TimeIcon,
  CheckmarkCircle as SuccessIcon,
  CloseCircle as FailIcon,
  Search as SearchIcon,
  Pulse as PulseIcon,
  ChatboxEllipses as ChatboxEllipsesIcon,
} from '@vicons/ionicons5'
import AddRouteModal from './components/AddRouteModal.vue'
import EditRouteModal from './components/EditRouteModal.vue'

// 注册 ECharts 组件
use([
  CanvasRenderer,
  LineChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
])

// 使用全局 API（不需要 provider）
const showMessage = (type, content) => {
  if (window.$message) {
    window.$message[type](content)
  } else {
    console.log(`[${type}] ${content}`)
  }
}

// i18n
const { t, locale } = useI18n()

// Language
const showLanguageModal = ref(false)
const currentLocale = ref(localStorage.getItem('app-locale') || 'zh-CN')

const switchLanguage = (lang) => {
  locale.value = lang
  currentLocale.value = lang
  localStorage.setItem('app-locale', lang)
  showLanguageModal.value = false
  showMessage("success", t('messages.languageChanged'))
}

// Page State
const currentPage = ref('home') // 'home' | 'models' | 'stats' | 'settings'
const refreshing = ref(false)
const compressing = ref(false)

// Theme
const isDark = ref(true)
const themeOverrides = {
  common: {
    primaryColor: '#18A058',
  },
}

const toggleTheme = () => {
  isDark.value = !isDark.value
  showMessage("info", isDark.value ? t('messages.switchedToDark') : t('messages.switchedToLight'))
}

// 刷新所有数据
const refreshAll = async () => {
  refreshing.value = true
  try {
    await Promise.all([
      loadRoutes(),
      loadStats(),
      loadConfig(),
      loadDailyStats(),
      loadHourlyStats(),
      loadSecondlyStats(),
      loadModelRanking(),
      loadUsageSummary()
    ])
    showMessage("success", t('messages.dataRefreshed'))
  } catch (error) {
    showMessage("error", t('messages.refreshFailed') + ': ' + error)
  } finally {
    refreshing.value = false
  }
}

// Settings
const settings = ref({
  redirectKeyword: 'proxy_auto',
  autoStart: false,
  minimizeToTray: false,
  enableFileLog: false,
  fallbackEnabled: true,
  tracesEnabled: false,
  tracesRetentionDays: 7,
  port: 5642,
})

const updateRedirectKeyword = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  try {
    await window.go.main.App.UpdateConfig(
      redirectConfig.value.enabled,
      settings.value.redirectKeyword,
      redirectConfig.value.targetModel,
      redirectConfig.value.targetRouteId
    )
    redirectConfig.value.keyword = settings.value.redirectKeyword
    showMessage("success", t('messages.redirectKeywordUpdated'))
    await loadConfig()
  } catch (error) {
    showMessage("error", t('messages.updateFailed') + ': ' + error)
  }
}

// 更新端口设置
const updatePort = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  try {
    await window.go.main.App.UpdatePort(settings.value.port)
    showMessage("success", t('settings.portUpdated'))
    // 提示用户需要重启
    showRestartDialog.value = true
  } catch (error) {
    showMessage("error", t('messages.updateFailed') + ': ' + error)
  }
}

// 重启应用
const restartApp = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  try {
    // 调用后端重启方法
    await window.go.main.App.RestartApp()
  } catch (error) {
    showMessage("error", t('messages.restartFailed') + ': ' + error)
  }
}

const saveSettings = () => {
  showMessage("info", t('messages.settingFailed'))
}

// 切换开机自启动
const toggleAutoStart = async (enabled) => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  try {
    await window.go.main.App.SetAutoStart(enabled)
    showMessage("success", enabled ? t('messages.autoStartEnabled') : t('messages.autoStartDisabled'))
  } catch (error) {
    showMessage("error", t('messages.settingFailed') + ': ' + error)
    settings.value.autoStart = !enabled // 恢复状态
  }
}

// 切换最小化到托盘
const toggleMinimizeToTray = async (enabled) => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  try {
    await window.go.main.App.SetMinimizeToTray(enabled)
    showMessage("success", enabled ? t('messages.minimizeEnabled') : t('messages.minimizeDisabled'))
  } catch (error) {
    showMessage("error", t('messages.settingFailed') + ': ' + error)
    settings.value.minimizeToTray = !enabled // 恢复状态
  }
}

// 切换文件日志
const toggleEnableFileLog = async (enabled) => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  try {
    await window.go.main.App.SetEnableFileLog(enabled)
    showMessage("success", enabled ? t('settings.fileLogEnabled') : t('settings.fileLogDisabled'))
  } catch (error) {
    showMessage("error", t('messages.settingFailed') + ': ' + error)
    settings.value.enableFileLog = !enabled // 恢复状态
  }
}

// 切换故障转移
const toggleFallbackEnabled = async (enabled) => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  try {
    await window.go.main.App.SetFallbackEnabled(enabled)
    showMessage("success", enabled ? t('settings.fallbackEnabled') : t('settings.fallbackDisabled'))
  } catch (error) {
    showMessage("error", t('messages.settingFailed') + ': ' + error)
    settings.value.fallbackEnabled = !enabled // 恢复状态
  }
}

// 切换 Traces 启用
const toggleTracesEnabled = async (enabled) => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  try {
    await window.go.main.App.SetTracesEnabled(enabled)
    showMessage("success", enabled ? t('settings.tracesEnabled') : t('settings.tracesDisabled'))
  } catch (error) {
    showMessage("error", t('messages.settingFailed') + ': ' + error)
    settings.value.tracesEnabled = !enabled
  }
}

// 更新 Traces 保留天数
const updateTracesRetentionDays = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    return
  }
  try {
    await window.go.main.App.SetTracesRetentionDays(settings.value.tracesRetentionDays)
  } catch (error) {
    showMessage("error", t('messages.settingFailed') + ': ' + error)
  }
}

// 压缩数据库
const compressDatabase = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  compressing.value = true
  try {
    const result = await window.go.main.App.CompressDatabase()
    const message = t('settings.compressResult', {
      before: result.before_count || 0,
      after: result.after_count || 0,
      deleted: result.deleted_logs || 0,
      hourly: result.hourly_stats_count || 0
    })
    showMessage("success", t('settings.compressSuccess') + ' - ' + message)
    // 刷新统计数据
    await loadStats()
    await loadDailyStats()
  } catch (error) {
    showMessage("error", t('settings.compressFailed') + ': ' + error)
  } finally {
    compressing.value = false
  }
}

// Stats
const stats = ref({
  route_count: 0,
  model_count: 0,
  total_requests: 0,
  total_tokens: 0,
  today_tokens: 0, // 今日token使用量
  today_requests: 0, // 今日请求数
  success_rate: 0,
})

// 热力图数据
const heatmapData = ref([])

// 热力图 tooltip 状态
const heatmapTooltip = ref({
  show: false,
  x: 0,
  y: 0,
  date: '',
  tokens: 0,
  requestTokens: 0,
  responseTokens: 0,
  requests: 0
})

// 显示热力图 tooltip（使用固定定位避免被边框遮挡）
const showHeatmapTooltip = (event, day) => {
  const rect = event.target.getBoundingClientRect()
  heatmapTooltip.value = {
    show: true,
    x: rect.left + rect.width / 2,
    y: rect.top,
    date: day.date,
    tokens: day.tokens,
    requestTokens: day.requestTokens || 0,
    responseTokens: day.responseTokens || 0,
    requests: day.requests
  }
}

// 生成热力图数据结构（填充空白日期）
const generateHeatmapData = (dailyStats) => {
  const weeks = []
  const today = new Date()
  const statsMap = {}

  // 将统计数据转换为map（包含 tokens 和 requests）
  if (dailyStats && Array.isArray(dailyStats)) {
    dailyStats.forEach(stat => {
      statsMap[stat.date] = {
        tokens: stat.total_tokens || 0,
        requestTokens: stat.request_tokens || 0,
        responseTokens: stat.response_tokens || 0,
        requests: stat.requests || 0
      }
    })
  }

  // 计算起始日期（52周前的周日）
  const startDate = new Date(today)
  startDate.setDate(startDate.getDate() - 363) // 回到约52周前
  // 调整到周日
  const dayOfWeek = startDate.getDay()
  startDate.setDate(startDate.getDate() - dayOfWeek)

  // 生成53周的数据（确保覆盖完整一年）
  for (let i = 0; i < 53; i++) {
    const week = []
    for (let j = 0; j < 7; j++) {
      const date = new Date(startDate)
      date.setDate(date.getDate() + (i * 7 + j))
      // 使用本地日期格式
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const dateStr = `${year}-${month}-${day}`
      const stat = statsMap[dateStr] || { tokens: 0, requestTokens: 0, responseTokens: 0, requests: 0 }
      week.push({
        date: dateStr,
        tokens: stat.tokens,
        requestTokens: stat.requestTokens,
        responseTokens: stat.responseTokens,
        requests: stat.requests
      })
    }
    weeks.push(week)
  }
  return weeks
}

// 动态计算月份标签（带位置信息）
const heatmapMonthsWithPosition = computed(() => {
  const monthsData = []
  const today = new Date()
  const startDate = new Date(today)
  startDate.setDate(startDate.getDate() - 363)
  // 调整到周日（与 generateHeatmapData 保持一致）
  const dayOfWeek = startDate.getDay()
  startDate.setDate(startDate.getDate() - dayOfWeek)
  
  // 使用 tm() 获取数组类型的翻译
  const monthNames = locale.value === 'zh-CN' 
    ? ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
    : ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  let lastMonth = -1
  
  // 遍历所有天数来检测月份变化
  for (let i = 0; i < 53; i++) {
    // 检查这一周的每一天，找到月份变化的位置
    for (let j = 0; j < 7; j++) {
      const date = new Date(startDate)
      date.setDate(date.getDate() + (i * 7 + j))
      const month = date.getMonth()
      if (month !== lastMonth) {
        // 如果是这周的第一天（周日）就是新月份开始，标记在这周
        // 否则标记在下一周
        const weekIndex = j === 0 ? i : (i < 52 ? i + 1 : i)
        // 避免重复添加同一个月
        if (monthsData.length === 0 || monthsData[monthsData.length - 1].name !== monthNames[month]) {
          monthsData.push({
            name: monthNames[month],
            weekIndex: j === 0 ? i : i
          })
        }
        lastMonth = month
        break // 找到这周的月份变化后跳出
      }
    }
  }
  return monthsData
})

const getHeatmapClass = (tokens) => {
  if (!tokens || tokens === 0) return 'level-0'
  if (tokens < 1000) return 'level-1'
  if (tokens < 5000) return 'level-2'
  if (tokens < 10000) return 'level-3'
  return 'level-4'
}

// 今日按小时统计数据
const hourlyStatsData = ref([])

// 秒级统计数据（今日全天）
const secondlyStatsData = ref([])

// 今日折线图配置（24小时 + 秒级数据点）
const todayChartOption = computed(() => {
  // 生成24小时的基础数据（填充空白小时为0）
  const hourlyMap = {}
  hourlyStatsData.value.forEach(stat => {
    hourlyMap[stat.hour] = {
      request_tokens: stat.request_tokens || 0,
      response_tokens: stat.response_tokens || 0,
      requests: stat.requests || 0
    }
  })

  // 24小时标签
  const hours = Array.from({ length: 24 }, (_, i) => `${String(i).padStart(2, '0')}:00`)
  
  // 24小时汇总数据（没有数据的小时为0）
  const hourlyRequestTokens = Array.from({ length: 24 }, (_, i) => hourlyMap[i]?.request_tokens || 0)
  const hourlyResponseTokens = Array.from({ length: 24 }, (_, i) => hourlyMap[i]?.response_tokens || 0)
  const hourlyRequests = Array.from({ length: 24 }, (_, i) => hourlyMap[i]?.requests || 0)

  // 处理秒级数据，转换为散点图数据（显示在对应小时位置）
  const secondlyData = secondlyStatsData.value || []
  const scatterRequestTokens = []
  const scatterResponseTokens = []
  const scatterRequests = []

  secondlyData.forEach(stat => {
    const ts = stat.timestamp || ''
    if (ts.length >= 19) {
      const hour = parseInt(ts.substring(11, 13), 10)
      const minute = parseInt(ts.substring(14, 16), 10)
      const second = parseInt(ts.substring(17, 19), 10)
      // 计算在X轴上的精确位置（小时 + 分钟/60 + 秒/3600）
      const xPos = hour + minute / 60 + second / 3600
      const timeLabel = ts.substring(11, 19)
      
      scatterRequestTokens.push({
        value: [xPos, stat.request_tokens || 0],
        time: timeLabel
      })
      scatterResponseTokens.push({
        value: [xPos, stat.response_tokens || 0],
        time: timeLabel
      })
      scatterRequests.push({
        value: [xPos, stat.requests || 0],
        time: timeLabel
      })
    }
  })

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      },
      formatter: function(params) {
        if (!params || params.length === 0) return ''
        
        let result = ''
        // 检查是否是散点数据
        const firstParam = params[0]
        if (firstParam.componentType === 'series' && firstParam.data && firstParam.data.time) {
          result = firstParam.data.time + '<br/>'
        } else {
          result = firstParam.axisValue + '<br/>'
        }
        
        params.forEach(param => {
          let value = param.data?.value?.[1] ?? param.value
          // 对 Token 数量进行格式化
          if (param.seriesName.includes('Token') || param.seriesName.includes('token')) {
            if (value >= 1000000) {
              value = (value / 1000000).toFixed(1) + 'M'
            } else if (value >= 1000) {
              value = (value / 1000).toFixed(1) + 'K'
            }
          }
          result += param.marker + param.seriesName + ': ' + value + '<br/>'
        })
        return result
      }
    },
    legend: {
      data: [
        t('stats.inputTokens') + '(' + t('stats.hourly') + ')',
        t('stats.outputTokens') + '(' + t('stats.hourly') + ')',
        t('stats.requestCount') + '(' + t('stats.hourly') + ')',
        t('stats.inputTokens') + '(' + t('stats.realtime') + ')',
        t('stats.outputTokens') + '(' + t('stats.realtime') + ')',
        t('stats.requestCount') + '(' + t('stats.realtime') + ')'
      ],
      textStyle: {
        color: isDark.value ? '#fff' : '#333'
      },
      type: 'scroll',
      bottom: 0
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: hours,
      axisLabel: {
        interval: 1
      }
    },
    yAxis: [
      {
        type: 'value',
        name: 'Tokens',
        position: 'left',
        axisLabel: {
          formatter: function(value) {
            if (value >= 1000000) {
              return (value / 1000000).toFixed(1) + 'M'
            } else if (value >= 1000) {
              return (value / 1000).toFixed(1) + 'K'
            }
            return value
          }
        }
      },
      {
        type: 'value',
        name: t('stats.requestCount'),
        position: 'right'
      }
    ],
    series: [
      // 小时汇总折线图
      {
        name: t('stats.inputTokens') + '(' + t('stats.hourly') + ')',
        type: 'line',
        smooth: true,
        data: hourlyRequestTokens,
        yAxisIndex: 0,
        areaStyle: {
          color: isDark.value ? 'rgba(24, 160, 88, 0.1)' : 'rgba(24, 160, 88, 0.2)'
        },
        lineStyle: {
          color: '#18a058',
          width: 2
        },
        itemStyle: {
          color: '#18a058'
        }
      },
      {
        name: t('stats.outputTokens') + '(' + t('stats.hourly') + ')',
        type: 'line',
        smooth: true,
        data: hourlyResponseTokens,
        yAxisIndex: 0,
        areaStyle: {
          color: isDark.value ? 'rgba(99, 125, 255, 0.1)' : 'rgba(99, 125, 255, 0.2)'
        },
        lineStyle: {
          color: '#637dff',
          width: 2
        },
        itemStyle: {
          color: '#637dff'
        }
      },
      {
        name: t('stats.requestCount') + '(' + t('stats.hourly') + ')',
        type: 'line',
        smooth: true,
        data: hourlyRequests,
        yAxisIndex: 1,
        lineStyle: {
          color: '#f0a020',
          width: 2
        },
        itemStyle: {
          color: '#f0a020'
        }
      },
      // 秒级散点图
      {
        name: t('stats.inputTokens') + '(' + t('stats.realtime') + ')',
        type: 'scatter',
        data: scatterRequestTokens,
        yAxisIndex: 0,
        symbolSize: 6,
        itemStyle: {
          color: '#18a058',
          opacity: 0.7
        }
      },
      {
        name: t('stats.outputTokens') + '(' + t('stats.realtime') + ')',
        type: 'scatter',
        data: scatterResponseTokens,
        yAxisIndex: 0,
        symbolSize: 6,
        itemStyle: {
          color: '#637dff',
          opacity: 0.7
        }
      },
      {
        name: t('stats.requestCount') + '(' + t('stats.realtime') + ')',
        type: 'scatter',
        data: scatterRequests,
        yAxisIndex: 1,
        symbolSize: 6,
        itemStyle: {
          color: '#f0a020',
          opacity: 0.7
        }
      }
    ]
  }
})

// 接口使用排行数据
const modelRankingData = ref([])

// 用量汇总数据
const usageSummary = ref({
  week_stats: [],
  year_stats: [],
  total_stats: {}
})

const rankingColumns = computed(() => [
  { title: t('stats.rank'), key: 'rank', width: 60 },
  {
    title: t('stats.model'),
    key: 'model',
    render(row) {
      return h(NTag, { type: 'info', size: 'small' }, { default: () => row.model })
    }
  },
  { title: t('stats.requests'), key: 'requests', width: 80 },
  {
    title: t('stats.inputTokens'),
    key: 'request_tokens',
    width: 100,
    render(row) {
      return formatNumber(row.request_tokens || 0)
    }
  },
  {
    title: t('stats.outputTokens'),
    key: 'response_tokens',
    width: 100,
    render(row) {
      return formatNumber(row.response_tokens || 0)
    }
  },
  {
    title: t('stats.totalTokensCol'),
    key: 'total_tokens',
    width: 100,
    render(row) {
      return formatNumber(row.total_tokens || 0)
    }
  },
  {
    title: t('stats.successRate'),
    key: 'success_rate',
    width: 80,
    render(row) {
      return `${row.success_rate || 0}%`
    }
  },
])

// 周用量表格列
const weeklyColumns = computed(() => [
  { title: t('stats.period'), key: 'period', width: 100 },
  { title: t('stats.requests'), key: 'request_count', width: 80 },
  {
    title: t('stats.totalTokensCol'),
    key: 'total_tokens',
    render(row) {
      return formatNumber(row.total_tokens || 0)
    }
  },
])

// 年用量表格列
const yearlyColumns = computed(() => [
  { title: t('stats.period'), key: 'period', width: 80 },
  { title: t('stats.requests'), key: 'request_count', width: 80 },
  {
    title: t('stats.totalTokensCol'),
    key: 'total_tokens',
    render(row) {
      return formatNumber(row.total_tokens || 0)
    }
  },
])

// ========== 请求日志相关 ==========
const logsData = ref([])
const logsLoading = ref(false)
const logsPage = ref(1)
const logsPageSize = ref(20)
const logsTotal = ref(0)
const logsFilter = ref({
  model: '',
  style: null,
  success: null,
})

// 筛选器选项
const styleOptions = computed(() => [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Claude', value: 'claude' },
  { label: 'Gemini', value: 'gemini' },
])

const successOptions = computed(() => [
  { label: t('logs.success'), value: 'true' },
  { label: t('logs.failed'), value: 'false' },
])

// 日志表格列定义
const logsColumns = computed(() => [
  {
    title: t('logs.time'),
    key: 'created_at',
    width: 160,
    render(row) {
      const date = new Date(row.created_at)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    }
  },
  {
    title: t('logs.model'),
    key: 'model',
    width: 180,
    ellipsis: { tooltip: true },
    render(row) {
      return h(NTag, { type: 'info', size: 'small' }, { default: () => row.model || '-' })
    }
  },
  {
    title: t('logs.providerModel'),
    key: 'provider_model',
    width: 180,
    ellipsis: { tooltip: true },
    render(row) {
      return row.provider_model || '-'
    }
  },
  {
    title: t('logs.provider'),
    key: 'provider_name',
    width: 120,
    render(row) {
      return row.provider_name || '-'
    }
  },
  {
    title: t('logs.style'),
    key: 'style',
    width: 80,
    render(row) {
      const styleMap = {
        'openai': { type: 'success', text: 'OpenAI' },
        'claude': { type: 'warning', text: 'Claude' },
        'gemini': { type: 'info', text: 'Gemini' },
      }
      const style = styleMap[row.style] || { type: 'default', text: row.style || '-' }
      return h(NTag, { type: style.type, size: 'small' }, { default: () => style.text })
    }
  },
  {
    title: t('logs.inputTokens'),
    key: 'request_tokens',
    width: 90,
    render(row) {
      return formatNumber(row.request_tokens || 0)
    }
  },
  {
    title: t('logs.outputTokens'),
    key: 'response_tokens',
    width: 90,
    render(row) {
      return formatNumber(row.response_tokens || 0)
    }
  },
  {
    title: t('logs.proxyTime'),
    key: 'proxy_time_ms',
    width: 90,
    render(row) {
      const ms = row.proxy_time_ms || 0
      if (ms >= 1000) {
        return (ms / 1000).toFixed(1) + 's'
      }
      return ms + 'ms'
    }
  },
  {
    title: t('logs.status'),
    key: 'success',
    width: 70,
    render(row) {
      if (row.success) {
        return h(NTag, { type: 'success', size: 'small' }, { default: () => t('logs.success') })
      } else {
        return h(
          NTooltip,
          { trigger: 'hover' },
          {
            trigger: () => h(NTag, { type: 'error', size: 'small' }, { default: () => t('logs.failed') }),
            default: () => row.error_message || t('logs.unknownError')
          }
        )
      }
    }
  },
  {
    title: t('logs.stream'),
    key: 'is_stream',
    width: 60,
    render(row) {
      return row.is_stream ? '✓' : '-'
    }
  },
])

// 加载请求日志
const loadRequestLogs = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  logsLoading.value = true
  try {
    // 通过 Wails v3 shim 调用后端服务
    const data = await window.go.main.App.GetRequestLogs(
      logsPage.value,
      logsPageSize.value,
      logsFilter.value.model || '',
      logsFilter.value.style || '',
      logsFilter.value.success || ''
    )
    logsData.value = data.data || []
    logsTotal.value = data.total || 0
  } catch (error) {
    console.error('加载请求日志失败:', error)
    showMessage("error", t('logs.loadFailed') + ': ' + error)
    logsData.value = []
    logsTotal.value = 0
  } finally {
    logsLoading.value = false
  }
}

// 防抖加载日志
let debounceTimer = null
const debounceLoadLogs = () => {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  debounceTimer = setTimeout(() => {
    logsPage.value = 1
    loadRequestLogs()
  }, 300)
}

// 处理每页数量变化
const handleLogsPageSizeChange = (pageSize) => {
  logsPageSize.value = pageSize
  logsPage.value = 1
  loadRequestLogs()
}

// 清空筛选器
const clearLogsFilter = () => {
  logsFilter.value = {
    model: '',
    style: null,
    success: null,
  }
  logsPage.value = 1
  loadRequestLogs()
}

// ========== 健康监控相关 ==========
const healthData = ref([])
const healthLoading = ref(false)

// 加载健康状态
const loadHealthStatus = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  healthLoading.value = true
  try {
    const data = await window.go.main.App.GetHealthStatus()
    healthData.value = data || []
  } catch (error) {
    console.error('加载健康状态失败:', error)
    showMessage("error", t('health.loadFailed') + ': ' + error)
    healthData.value = []
  } finally {
    healthLoading.value = false
  }
}

// ========== Traces 对话追踪相关 ==========
const allTraces = ref([])
const allTracesPage = ref(1)
const allTracesPageSize = ref(20)
const allTracesTotal = ref(0)
const tracesLoading = ref(false)
const tracesSearchQuery = ref('')
const tracesAutoRefresh = ref(false)
let tracesAutoRefreshTimer = null
let tracesSearchTimer = null

// 过滤后的 traces
const filteredTraces = computed(() => {
  if (!tracesSearchQuery.value) {
    return allTraces.value
  }
  const query = tracesSearchQuery.value.toLowerCase()
  return allTraces.value.filter(trace => 
    trace.model?.toLowerCase().includes(query) ||
    trace.provider_name?.toLowerCase().includes(query) ||
    trace.remote_ip?.includes(query) ||
    trace.request_content?.toLowerCase().includes(query) ||
    trace.response_content?.toLowerCase().includes(query)
  )
})

// 加载所有 trace 记录
const loadAllTraces = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", t('messages.wailsNotReady'))
    return
  }
  tracesLoading.value = true
  try {
    const data = await window.go.main.App.GetAllTraces(allTracesPage.value, allTracesPageSize.value)
    allTraces.value = data.traces || []
    allTracesTotal.value = data.total || 0
  } catch (error) {
    console.error('加载 Traces 失败:', error)
    showMessage("error", t('traces.loadFailed') + ': ' + error)
    allTraces.value = []
  } finally {
    tracesLoading.value = false
  }
}

// 防抖搜索
const debounceSearchTraces = () => {
  if (tracesSearchTimer) {
    clearTimeout(tracesSearchTimer)
  }
  tracesSearchTimer = setTimeout(() => {
    // 搜索已经通过 computed 实现，不需要额外操作
  }, 300)
}

// 切换自动刷新
const toggleTracesAutoRefresh = (enabled) => {
  if (enabled) {
    tracesAutoRefreshTimer = setInterval(() => {
      loadAllTraces()
    }, 5000)
  } else {
    if (tracesAutoRefreshTimer) {
      clearInterval(tracesAutoRefreshTimer)
      tracesAutoRefreshTimer = null
    }
  }
}

// 处理每页数量变化
const handleTracesPageSizeChange = (pageSize) => {
  allTracesPageSize.value = pageSize
  allTracesPage.value = 1
  loadAllTraces()
}

// 清除所有 Traces
const clearAllTraces = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    return
  }
  try {
    const deleted = await window.go.main.App.ClearAllTraces()
    showMessage("success", t('traces.cleared') + `: ${deleted}`)
    allTraces.value = []
    allTracesTotal.value = 0
    await loadAllTraces()
  } catch (error) {
    showMessage("error", t('traces.clearFailed') + ': ' + error)
  }
}

// 格式化 JSON
const formatJson = (str) => {
  if (!str) return ''
  try {
    const obj = JSON.parse(str)
    return JSON.stringify(obj, null, 2)
  } catch (e) {
    return str
  }
}

// Config
const config = ref({
  localApiKey: '',
  localApiEndpoint: '',
})

// Redirect Config
const redirectConfig = ref({
  enabled: false,
  keyword: 'proxy_auto',
  targetModel: '',
  targetName: '',
  targetRouteId: 0,
})

// Routes
const routes = ref([])
const showAddModal = ref(false)
const showEditModal = ref(false)
const editingRoute = ref(null)
const expandedGroups = ref([]) // 控制折叠面板展开状态
const fileInput = ref(null) // 文件输入引用
const showClearDialog = ref(false) // 清除数据确认对话框
const showRestartDialog = ref(false) // 重启确认对话框
const showEditApiKeyModal = ref(false) // 编辑 API Key 对话框
const newApiKey = ref('') // 新 API Key 输入值

// Computed: 按分组组织路由
const groupedRoutes = computed(() => {
  const groups = {}
  routes.value.forEach(route => {
    const groupName = route.group || '未分组'
    if (!groups[groupName]) {
      groups[groupName] = []
    }
    groups[groupName].push(route)
  })
  return groups
})


// 行属性设置
const rowProps = (row) => {
  return {
    'data-model': row.model
  }
}

// Pagination
const pagination = {
  pageSize: 10,
}

// 设置为重定向按钮处理
const setAsRedirect = async (row) => {
  redirectConfig.value.targetModel = row.model
  redirectConfig.value.targetRouteId = row.id
  redirectConfig.value.enabled = true
  await saveRedirectConfig()
  showMessage("success", t('home.setRedirectSuccess'))
}

// 跳转到目标模型
const jumpToTargetModel = () => {
  currentPage.value = 'models'

  // 展开所有分组
  expandedGroups.value = Object.keys(groupedRoutes.value)

  // 等待DOM更新后滚动到目标模型
  nextTick(() => {
    // 查找目标模型所在的行
    const targetRows = document.querySelectorAll('[data-model="' + redirectConfig.value.targetModel + '"]')
    if (targetRows.length > 0) {
      targetRows[0].scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  })
}

// Table columns for home page
const columns = [
  {
    title: 'ID',
    key: 'id',
    width: 60,
  },
  {
    title: '名称',
    key: 'name',
    width: 150,
  },
  {
    title: '模型',
    key: 'model',
    width: 180,
    render(row) {
      return h(NTag, { type: 'info' }, { default: () => row.model })
    },
  },
  {
    title: 'API URL',
    key: 'api_url',
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: 'API Key',
    key: 'api_key',
    width: 150,
    render(row) {
      return maskApiKey(row.api_key)
    },
  },
  {
    title: '分组',
    key: 'group',
    width: 100,
    render(row) {
      return row.group ? h(NTag, { type: 'success', size: 'small' }, { default: () => row.group }) : '-'
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render(row) {
      return h(NSpace, {}, {
        default: () => [
          h(
            NButton,
            {
              size: 'small',
              onClick: () => handleEdit(row),
            },
            { default: () => '编辑', icon: () => h(NIcon, {}, { default: () => h(EditIcon) }) }
          ),
          h(
            NButton,
            {
              size: 'small',
              type: 'error',
              onClick: () => handleDelete(row),
            },
            { default: () => '删除', icon: () => h(NIcon, {}, { default: () => h(DeleteIcon) }) }
          ),
        ]
      })
    },
  },
]

// Table columns for models page (with redirect button)
const modelsPageColumns = computed(() => [
  {
    title: 'ID',
    key: 'id',
    width: 60,
  },
  {
    title: t('models.name'),
    key: 'name',
    width: 150,
  },
  {
    title: t('models.model'),
    key: 'model',
    width: 200,
    render(row) {
      return h(NSpace, { align: 'center', size: 'small' }, {
        default: () => [
          h(NTag, { type: 'info', size: 'small' }, { default: () => row.model }),
          // 如果是当前重定向目标，通过路由ID精确匹配（避免同ID跨分组显示问题）
          (redirectConfig.value.targetRouteId === row.id || 
           (redirectConfig.value.targetRouteId === 0 && redirectConfig.value.targetModel === row.model))
            ? h(NTag, { type: 'success', size: 'small' }, { default: () => t('home.redirectTarget') })
            : null
        ]
      })
    },
  },
  {
    title: t('models.apiUrl'),
    key: 'api_url',
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: t('models.enabled'),
    key: 'enabled',
    width: 100,
    render(row) {
      return h(NSpace, { align: 'center' }, {
        default: () => [
          h(NSwitch, {
            value: row.enabled,
            onUpdateValue: (val) => handleToggleRoute(row.id, val),
          }),
          h('span', { style: { fontSize: '12px', color: row.enabled ? '#18a058' : '#999' } }, 
            row.enabled ? t('models.enabledStatus') : t('models.disabledStatus')
          )
        ]
      })
    },
  },
  {
    title: t('models.actions'),
    key: 'actions',
    width: 280,
    render(row) {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(
            NButton,
            {
              size: 'small',
              onClick: () => handleEdit(row),
            },
            { default: () => t('models.edit'), icon: () => h(NIcon, { size: 14 }, { default: () => h(EditIcon) }) }
          ),
          h(
            NButton,
            {
              size: 'small',
              type: 'error',
              onClick: () => handleDelete(row),
            },
            { default: () => t('models.delete'), icon: () => h(NIcon, { size: 14 }, { default: () => h(DeleteIcon) }) }
          ),
          h(
            NButton,
            {
              size: 'small',
              type: 'primary',
              onClick: () => setAsRedirect(row),
            },
            { default: () => t('models.setAsTarget'), icon: () => h(NIcon, { size: 14 }, { default: () => h(LinkIcon) }) }
          ),
        ]
      })
    },
  },
])

// Computed
const modelOptions = computed(() => {
  const models = routes.value.map(r => r.model)
  const uniqueModels = [...new Set(models)]
  return uniqueModels.map(m => ({ label: m, value: m }))
})

// Methods
const loadRoutes = async () => {
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      console.error('Wails runtime not available')
      return
    }
    const data = await window.go.main.App.GetRoutes()
    routes.value = data || []
    console.log('Routes loaded:', routes.value.length)

    // 自动展开所有分组
    expandedGroups.value = Object.keys(groupedRoutes.value)
  } catch (error) {
    console.error('Failed to load routes:', error)
    showMessage("error", t('messages.refreshFailed') + ': ' + error)
  }
}

const loadStats = async () => {
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      console.error('Wails runtime not available')
      return
    }
    const data = await window.go.main.App.GetStats()
    stats.value = data || stats.value
    console.log('Stats loaded:', stats.value)
  } catch (error) {
    console.error('加载统计失败:', error)
  }
}

// 加载每日统计（用于热力图）
const loadDailyStats = async () => {
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      return
    }
    const data = await window.go.main.App.GetDailyStats(365) // 获取365天数据
    heatmapData.value = generateHeatmapData(data || [])
  } catch (error) {
    console.error('加载每日统计失败:', error)
  }
}

// 加载今日按小时统计（用于折线图）
const loadHourlyStats = async () => {
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      return
    }
    const data = await window.go.main.App.GetHourlyStats()
    hourlyStatsData.value = data || []
  } catch (error) {
    console.error('加载按小时统计失败:', error)
  }
}

// 加载秒级统计（用于实时折线图）
const loadSecondlyStats = async () => {
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      return
    }
    // 获取最近60分钟的秒级数据
    const data = await window.go.main.App.GetSecondlyStats(60)
    secondlyStatsData.value = data || []
  } catch (error) {
    console.error('加载秒级统计失败:', error)
  }
}

// 加载模型使用排行
const loadModelRanking = async () => {
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      return
    }
    const data = await window.go.main.App.GetModelRanking(10) // 获取前10名
    modelRankingData.value = data || []
  } catch (error) {
    console.error('加载模型排行失败:', error)
  }
}

// 加载用量汇总
const loadUsageSummary = async () => {
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      return
    }
    const data = await window.go.main.App.GetUsageSummary()
    usageSummary.value = data || { week_stats: [], year_stats: [], total_stats: {} }
  } catch (error) {
    console.error('加载用量汇总失败:', error)
  }
}

const loadConfig = async () => {
  try {
    if (!window.go || !window.go.main || !window.go.main.App) {
      console.error('Wails runtime not available')
      return
    }
    const data = await window.go.main.App.GetConfig()
    // 映射后端字段名到前端字段名
    config.value = {
      localApiKey: data.localApiKey || '',
      localApiEndpoint: data.openaiEndpoint || ''
    }
    redirectConfig.value.enabled = data.redirectEnabled || false
    redirectConfig.value.keyword = data.redirectKeyword || 'proxy_auto'
    redirectConfig.value.targetModel = data.redirectTargetModel || ''
    redirectConfig.value.targetName = data.redirectTargetName || ''
    redirectConfig.value.targetRouteId = data.redirectTargetRouteId || 0
    settings.value.redirectKeyword = data.redirectKeyword || 'proxy_auto' // 同步到设置
    settings.value.minimizeToTray = data.minimizeToTray || false
    settings.value.autoStart = data.autoStart || false
    settings.value.enableFileLog = data.enableFileLog || false
    settings.value.fallbackEnabled = data.fallbackEnabled !== false // 默认启用
    settings.value.tracesEnabled = data.tracesEnabled || false
    settings.value.tracesRetentionDays = data.tracesRetentionDays || 7
    settings.value.port = data.port || 5642
    console.log('Config loaded:', config.value)
  } catch (error) {
    console.error('加载配置失败:', error)
  }
}

const saveRedirectConfig = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", 'Wails 运行时未就绪')
    return
  }
  try {
    await window.go.main.App.UpdateConfig(
      redirectConfig.value.enabled,
      redirectConfig.value.keyword,
      redirectConfig.value.targetModel,
      redirectConfig.value.targetRouteId
    )
    showMessage("success", t('messages.redirectConfigSaved'))
    // 重新加载配置以获取最新的 targetName
    await loadConfig()
  } catch (error) {
    showMessage("error", t('messages.redirectConfigFailed') + ': ' + error)
  }
}

// 清理 API URL，移除末尾斜杠
const handleRouteAdded = () => {
  loadRoutes()
  loadStats()
}

const handleRouteUpdated = () => {
  loadRoutes()
  loadStats()
}

const handleEdit = (row) => {
  editingRoute.value = row
  showEditModal.value = true
}

const handleDelete = async (row) => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", 'Wails 运行时未就绪')
    return
  }
  try {
    await window.go.main.App.DeleteRoute(row.id)
    showMessage("success", t('deleteRoute.deleted'))
    loadRoutes()
    loadStats()
  } catch (error) {
    showMessage("error", t('deleteRoute.deleteFailed') + ': ' + error)
  }
}

// 启用/禁用路由
const handleToggleRoute = async (id, enabled) => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", 'Wails 运行时未就绪')
    return
  }
  try {
    await window.go.main.App.ToggleRoute(id, enabled)
    showMessage("success", enabled ? t('models.routeEnabled') : t('models.routeDisabled'))
    loadRoutes()
    loadStats()
  } catch (error) {
    showMessage("error", t('messages.updateFailed') + ': ' + error)
  }
}



const maskApiKey = (key) => {
  if (!key || key.length <= 10) return key
  return key.substring(0, 5) + '***' + key.substring(key.length - 5)
}

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    showMessage("success", t('messages.copySuccess'))
  } catch (error) {
    showMessage("error", t('messages.copyFailed'))
  }
}

const formatNumber = (num) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

// 生成随机 API Key
const generateRandomApiKey = () => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = 'sk-'
  for (let i = 0; i < 48; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

// 随机更新 API Key
const generateNewApiKey = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", 'Wails 运行时未就绪')
    return
  }

  try {
    const randomKey = generateRandomApiKey()
    await window.go.main.App.UpdateLocalApiKey(randomKey)
    showMessage("success", t('home.apiKeyRandomized'))
    await loadConfig() // 重新加载配置
  } catch (error) {
    showMessage("error", t('messages.updateFailed') + ': ' + error)
  }
}

// 保存自定义 API Key
const saveCustomApiKey = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", 'Wails 运行时未就绪')
    return
  }

  if (!newApiKey.value.trim()) {
    showMessage("warning", t('home.apiKeyRequired'))
    return
  }

  try {
    await window.go.main.App.UpdateLocalApiKey(newApiKey.value.trim())
    showMessage("success", t('home.apiKeySaved'))
    showEditApiKeyModal.value = false
    newApiKey.value = ''
    await loadConfig() // 重新加载配置
  } catch (error) {
    showMessage("error", t('messages.updateFailed') + ': ' + error)
  }
}

// 导出路由为 JSON
const exportRoutes = () => {
  try {
    const exportData = routes.value.map(route => ({
      name: route.name,
      model: route.model,
      api_url: route.api_url,
      api_key: route.api_key,
      group: route.group,
    }))

    const jsonStr = JSON.stringify(exportData, null, 2)
    const blob = new Blob([jsonStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `openai-router-routes-${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)

    showMessage("success", t('models.exportSuccess'))
  } catch (error) {
    showMessage("error", t('models.exportFailed') + ': ' + error)
  }
}

// 触发文件选择
const triggerImport = () => {
  fileInput.value?.click()
}

// 显示清除数据确认对话框
const showClearStatsDialog = () => {
  showClearDialog.value = true
}

// 确认清除统计数据
const confirmClearStats = async () => {
  if (!window.go || !window.go.main || !window.go.main.App) {
    showMessage("error", 'Wails 运行时未就绪')
    return
  }

  try {
    await window.go.main.App.ClearStats()
    showMessage("success", t('stats.clearSuccess'))
    showClearDialog.value = false

    // 重新加载数据
    await loadStats()
    await loadDailyStats()
    await loadHourlyStats()
    await loadSecondlyStats()
    await loadModelRanking()
  } catch (error) {
    showMessage("error", t('stats.clearFailed') + ': ' + error)
  }
}

// 处理文件导入
const handleFileImport = async (event) => {
  const file = event.target.files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const importData = JSON.parse(text)

    if (!Array.isArray(importData)) {
      showMessage("error", 'JSON 格式错误：应为路由数组')
      return
    }

    if (!window.go || !window.go.main || !window.go.main.App) {
      showMessage("error", 'Wails 运行时未就绪')
      return
    }

    let successCount = 0
    let failCount = 0

    for (const route of importData) {
      try {
        await window.go.main.App.AddRoute(
          route.name || '',
          route.model || '',
          route.api_url || '',
          route.api_key || '',
          route.group || ''
        )
        successCount++
      } catch (error) {
        console.error('导入路由失败:', route, error)
        failCount++
      }
    }

    showMessage("success", t('models.importSuccess', { count: successCount }))
    loadRoutes()
    loadStats()
  } catch (error) {
    showMessage("error", t('models.importFailed') + ': ' + error)
  } finally {
    // 清空文件输入
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  }
}

// Lifecycle
onMounted(async () => {
  // Wait for Wails runtime to be ready
  if (!window.go) {
    console.log('Waiting for Wails runtime...')
    await new Promise((resolve) => {
      const checkRuntime = setInterval(() => {
        if (window.go) {
          clearInterval(checkRuntime)
          resolve()
        }
      }, 100)
    })
  }

  console.log('Wails runtime ready, loading data...')
  loadRoutes()
  loadStats()
  loadConfig()
  loadDailyStats()
  loadHourlyStats()
  loadSecondlyStats()
  loadModelRanking()
  loadUsageSummary()

  // 每 10 秒刷新一次秒级统计（实时数据）
  setInterval(() => {
    loadSecondlyStats()
  }, 10000)

  // 每 30 秒刷新一次统计
  setInterval(() => {
    loadStats()
    loadHourlyStats()
  }, 30000)

  // 每 5 分钟刷新一次热力图和排行
  setInterval(() => {
    loadDailyStats()
    loadModelRanking()
    loadUsageSummary()
  }, 300000)
})

// Watch groupedRoutes to automatically expand all groups when they change
watch(groupedRoutes, (newGroups) => {
  console.log('Grouped routes changed, expanding all groups')
  expandedGroups.value = Object.keys(newGroups)
}, { deep: true })

// Watch currentPage to load logs when switching to logs page
watch(currentPage, (newPage) => {
  if (newPage === 'logs') {
    loadRequestLogs()
  }
})
</script>

<style>
/* 全局滚动条隐藏 - Wails 专用 */
:deep(*)::-webkit-scrollbar {
  width: 0px !important;
  height: 0px !important;
  background: transparent !important;
  display: none !important;
}

:deep(*) {
  scrollbar-width: none !important;
  -ms-overflow-style: none !important;
}

/* 针对 Naive UI 组件的特殊处理 */
:deep(.n-layout-content) {
  overflow-y: auto !important;
  overflow-x: hidden !important;
}

:deep(.n-layout-content::-webkit-scrollbar),
:deep(.n-data-table::-webkit-scrollbar),
:deep(.n-card::-webkit-scrollbar),
:deep(.n-scrollbar::-webkit-scrollbar),
:deep(.n-collapse-item::-webkit-scrollbar),
:deep(.n-tab-pane::-webkit-scrollbar) {
  width: 0px !important;
  height: 0px !important;
  background: transparent !important;
  display: none !important;
}

/* 覆盖 Naive UI n-input 组件的 height */
.n-input .n-input__input-el {
  height: auto !important;
}
</style>

<style scoped>
:deep(.n-card__content) {
  padding: 16px;
}

:deep(.n-statistic) {
  color: white;
}

:deep(.n-statistic .n-statistic__label) {
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
}

:deep(.n-statistic .n-statistic__value) {
  color: white;
  font-size: 28px;
  font-weight: 600;
}



/* GitHub 热力图样式 - 全屏版本 */
.heatmap-container {
  padding: 20px;
  position: relative;
  width: 100%;
  overflow-x: auto;
}

.heatmap-months-row {
  position: relative;
  height: 20px;
  margin-bottom: 8px;
  font-size: 12px;
  color: #888;
  width: 100%;
}

.heatmap-month-label {
  position: absolute;
  white-space: nowrap;
  transform: translateX(0);
}

.heatmap-grid {
  display: flex;
  gap: 4px;
  margin-bottom: 12px;
  width: 100%;
  justify-content: flex-start;
  overflow-x: auto;
}

.heatmap-week {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 0 0 auto;
  width: calc((100% - 52 * 4px) / 53);
  min-width: 12px;
}

.heatmap-cell {
  width: 100%;
  aspect-ratio: 1;
  min-width: 10px;
  max-width: 16px;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.2s;
}

.heatmap-cell:hover {
  transform: scale(1.5);
  border: 1px solid #fff;
  z-index: 10;
}

.heatmap-cell.level-0 {
  background-color: #3a3a3a;
}

.heatmap-cell.level-1 {
  background-color: #9be9a8;
}

.heatmap-cell.level-2 {
  background-color: #40c463;
}

.heatmap-cell.level-3 {
  background-color: #30a14e;
}

.heatmap-cell.level-4 {
  background-color: #216e39;
}

.heatmap-tooltip {
  position: fixed;
  background: rgba(0, 0, 0, 0.85);
  color: #fff;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 12px;
  pointer-events: none;
  z-index: 1000;
  white-space: nowrap;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  transform: translate(-50%, -100%);
  margin-top: -10px;
}

.heatmap-legend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #888;
  justify-content: flex-end;
  margin-top: 8px;
}

.legend-box {
  width: 11px;
  height: 11px;
  border-radius: 2px;
}

.legend-box.level-0 {
  background-color: #3a3a3a;
}

.legend-box.level-1 {
  background-color: #9be9a8;
}

.legend-box.level-2 {
  background-color: #40c463;
}

.legend-box.level-3 {
  background-color: #30a14e;
}

.legend-box.level-4 {
  background-color: #216e39;
}

/* 健康监控状态条样式 */
.health-route-item {
  padding: 8px 12px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.02);
  transition: background 0.2s;
}

.health-route-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.status-bar-container {
  flex: 1;
  max-width: 500px;
  min-width: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-bar {
  display: flex;
  gap: 2px;
  align-items: center;
}

.status-dot {
  width: 8px;
  height: 20px;
  border-radius: 2px;
  transition: transform 0.15s;
}

.status-dot:hover {
  transform: scaleY(1.3);
}

.status-dot.success {
  background-color: #18a058;
}

.status-dot.fail {
  background-color: #d03050;
}
</style>
