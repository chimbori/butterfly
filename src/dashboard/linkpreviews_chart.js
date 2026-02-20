/**
 * @fileoverview Renders charts for link preview statistics on the dashboard.
 */
export function initLinkPreviewsCharts() {
  function init() {
    initDomainChart();
    initUserAgentChart();
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
}

async function initDomainChart() {
  const canvas = document.getElementById('linkpreviews-domain-chart');
  if (!canvas) {
    return;
  }

  try {
    const response = await fetch('/dashboard/link-previews/stats');
    if (!response.ok) {
      console.error('Failed to fetch link preview statistics');
      return;
    }

    const stats = await response.json();
    if (stats.length === 0) {
      console.log('No link preview data available');
      return;
    }

    new Chart(canvas.getContext('2d'), {
      type: 'doughnut',
      data: {
        labels: stats.map(d => d.Domain),
        datasets: [{
          data: stats.map(d => d.TotalAccesses),
          borderWidth: 1
        }]
      },
      options: {
        responsive: true,
        plugins: {
          legend: {
            position: 'right',
            labels: {
              font: {
                size: 14,
                family: 'Inter'
              }
            }
          }
        }
      }
    });
  } catch (error) {
    console.error('Error loading link preview chart:', error);
  }
}

function initUserAgentChart() {
  const canvas = document.getElementById('linkpreviews-useragents-chart');
  if (!canvas) {
    return;
  }

  const rangeContainer = document.getElementById('linkpreviews-useragents-range');
  const rangeButtons = rangeContainer ? rangeContainer.querySelectorAll('button[data-days]') : [];

  let chart = null;
  let currentDays = 7;

  function setActiveButton(days) {
    rangeButtons.forEach(button => {
      const isActive = Number(button.dataset.days) === days;
      button.setAttribute('aria-pressed', isActive ? 'true' : 'false');
    });
  }

  function buildDatasets(stats) {
    const labelSet = new Set();
    const agentSet = new Set();
    const totalsByAgent = new Map();

    stats.forEach(row => {
      labelSet.add(row.Day);
      agentSet.add(row.CanonicalUserAgent);
      totalsByAgent.set(
        row.CanonicalUserAgent,
        (totalsByAgent.get(row.CanonicalUserAgent) || 0) + row.TotalAccesses
      );
    });

    const labels = Array.from(labelSet).sort((a, b) => b.localeCompare(a));
    const agents = Array.from(agentSet).sort((a, b) => (totalsByAgent.get(b) || 0) - (totalsByAgent.get(a) || 0));
    const matrix = new Map();

    stats.forEach(row => {
      const key = `${row.Day}::${row.CanonicalUserAgent}`;
      matrix.set(key, row.TotalAccesses);
    });

    const datasets = agents.map(agent => ({
      label: agent,
      data: labels.map(day => matrix.get(`${day}::${agent}`) || 0)
    }));

    return { labels, datasets };
  }

  async function loadChart(days) {
    try {
      const response = await fetch(`/dashboard/link-previews/user-agents?days=${days}`);
      if (!response.ok) {
        console.error('Failed to fetch link preview user agent statistics');
        return;
      }

      const stats = await response.json();
      if (stats.length === 0) {
        console.log('No user agent data available');
        return;
      }

      const { labels, datasets } = buildDatasets(stats);

      if (chart) {
        chart.data.labels = labels;
        chart.data.datasets = datasets;
        chart.update();
        return;
      }

      chart = new Chart(canvas.getContext('2d'), {
        type: 'bar',
        data: {
          labels,
          datasets
        },
        options: {
          indexAxis: 'y',
          responsive: true,
          scales: {
            x: {
              stacked: true,
              beginAtZero: true
            },
            y: {
              stacked: true
            }
          },
          plugins: {
            legend: {
              position: 'right'
            }
          }
        }
      });
    } catch (error) {
      console.error('Error loading user agent chart:', error);
    }
  }

  if (rangeButtons.length > 0) {
    rangeButtons.forEach(button => {
      button.addEventListener('click', () => {
        const days = Number(button.dataset.days || 7);
        if (!Number.isNaN(days)) {
          currentDays = days;
          setActiveButton(days);
          loadChart(days);
        }
      });
    });
  }

  setActiveButton(currentDays);
  loadChart(currentDays);
}
