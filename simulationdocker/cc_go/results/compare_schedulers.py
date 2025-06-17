#!/usr/bin/env python3

import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
import sys
import os

def load_results(filepath):
    return pd.read_csv(filepath, parse_dates=['Timestamp'])

def compute_metrics(df):
    metrics = {
        'total_containers': len(df),
        'success_rate': df['Success'].mean() * 100,
        'avg_latency': df[df['Success'] == True]['SchedulingLatency(ms)'].mean(),
        'avg_utilization': df[df['Success'] == True]['ResourceUtilization'].mean() * 100,
        'p95_latency': df[df['Success'] == True]['SchedulingLatency(ms)'].quantile(0.95),
        'p99_latency': df[df['Success'] == True]['SchedulingLatency(ms)'].quantile(0.99)
    }
    
    # Calculate container type distribution
    type_dist = df[df['Success'] == True]['ContainerType'].value_counts(normalize=True) * 100
    for ctype, percentage in type_dist.items():
        metrics[f'type_{ctype}_pct'] = percentage
    
    return metrics

def compare_schedulers(results_dir):
    files = [f for f in os.listdir(results_dir) if f.endswith('_results.csv')]
    if not files:
        print(f"No result files found in {results_dir}")
        return
    
    all_metrics = {}
    all_dfs = {}
    
    for file in files:
        scheduler_name = file.split('_')[0].capitalize()
        filepath = os.path.join(results_dir, file)
        df = load_results(filepath)
        all_dfs[scheduler_name] = df
        all_metrics[scheduler_name] = compute_metrics(df)
    
    # Create comparison tables
    metrics_df = pd.DataFrame(all_metrics).T
    print("\n=== Scheduler Performance Comparison ===")
    print(metrics_df[['total_containers', 'success_rate', 'avg_latency', 'avg_utilization']].round(2))
    print("\n=== Detailed Latency Metrics (ms) ===")
    print(metrics_df[['avg_latency', 'p95_latency', 'p99_latency']].round(2))
    
    # Generate plots
    create_plots(all_dfs, all_metrics, results_dir)

def create_plots(all_dfs, all_metrics, results_dir):
    sns.set_style("whitegrid")
    
    # Create figures directory if it doesn't exist
    figures_dir = os.path.join(results_dir, 'figures')
    os.makedirs(figures_dir, exist_ok=True)
    
    # 1. Utilization over time
    plt.figure(figsize=(12, 6))
    for name, df in all_dfs.items():
        # Group by time buckets and calculate mean utilization
        df['time_bucket'] = pd.cut(df['Timestamp'], 20)
        util_over_time = df.groupby('time_bucket')['ResourceUtilization'].mean()
        plt.plot(util_over_time.index.astype(str), util_over_time.values, 'o-', label=name)
    
    plt.xlabel('Time')
    plt.ylabel('Resource Utilization')
    plt.title('Resource Utilization Over Time')
    plt.legend()
    plt.xticks(rotation=45)
    plt.tight_layout()
    plt.savefig(os.path.join(figures_dir, 'utilization_over_time.png'))
    
    # 2. Success rate by container type
    plt.figure(figsize=(12, 6))
    success_by_type = {}
    
    for name, df in all_dfs.items():
        success_rate = df.groupby('ContainerType')['Success'].mean() * 100
        success_by_type[name] = success_rate
    
    success_df = pd.DataFrame(success_by_type)
    success_df.plot(kind='bar', figsize=(12, 6))
    plt.xlabel('Container Type')
    plt.ylabel('Success Rate (%)')
    plt.title('Scheduling Success Rate by Container Type')
    plt.legend(title='Scheduler')
    plt.tight_layout()
    plt.savefig(os.path.join(figures_dir, 'success_by_type.png'))
    
    # 3. Scheduling latency comparison
    plt.figure(figsize=(12, 6))
    latency_data = []
    
    for name, df in all_dfs.items():
        successful_df = df[df['Success'] == True]
        latency_data.append(successful_df['SchedulingLatency(ms)'])
    
    plt.boxplot(latency_data, labels=all_dfs.keys())
    plt.xlabel('Scheduler')
    plt.ylabel('Scheduling Latency (ms)')
    plt.title('Scheduling Latency Comparison')
    plt.tight_layout()
    plt.savefig(os.path.join(figures_dir, 'latency_comparison.png'))
    
    # 4. Resource utilization comparison
    plt.figure(figsize=(10, 6))
    utilization_data = [all_metrics[name]['avg_utilization'] for name in all_dfs.keys()]
    colors = ['#3498db', '#2ecc71', '#e74c3c']
    
    bars = plt.bar(all_dfs.keys(), utilization_data, color=colors)
    
    plt.xlabel('Scheduler')
    plt.ylabel('Average Resource Utilization (%)')
    plt.title('Resource Utilization by Scheduler')
    
    # Add value labels on top of bars
    for bar in bars:
        height = bar.get_height()
        plt.text(bar.get_x() + bar.get_width()/2., height + 0.5,
                f'{height:.1f}%',
                ha='center', va='bottom')
    
    plt.tight_layout()
    plt.savefig(os.path.join(figures_dir, 'utilization_comparison.png'))
    
    print(f"\nPlots saved to {figures_dir}")

if __name__ == "__main__":
    results_dir = 'results'
    if len(sys.argv) > 1:
        results_dir = sys.argv[1]
    
    compare_schedulers(results_dir)
