/**
 * @fileoverview Renders a doughnut chart for link preview statistics on the dashboard.
 */
export function initLinkPreviewsChart() {
  async function loadChart() {
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

      new Chart(document.getElementById('linkpreviews-domain-chart').getContext('2d'), {
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

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', loadChart);
  } else {
    loadChart();
  }
}
