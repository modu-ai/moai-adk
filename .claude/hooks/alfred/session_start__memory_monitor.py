#!/usr/bin/env python3
"""
Memory Monitor Hook for Claude Code Sessions
Monitors memory usage and provides alerts for potential memory leaks
"""
import psutil
import json
import os
import sys
from datetime import datetime

def check_memory_usage():
    """Check current memory usage and return status"""
    try:
        memory = psutil.virtual_memory()
        process = psutil.Process()

        # Get system-wide memory info
        total_memory = memory.total
        used_memory = memory.used
        available_memory = memory.available
        memory_percent = memory.percent

        # Get current process memory info
        process_memory = process.memory_info()
        process_rss = process_memory.rss
        process_vms = process_memory.vms

        # Convert to GB for readability
        total_gb = total_memory / (1024**3)
        used_gb = used_memory / (1024**3)
        available_gb = available_memory / (1024**3)
        process_rss_gb = process_rss / (1024**3)

        return {
            "timestamp": datetime.now().isoformat(),
            "system": {
                "total_gb": round(total_gb, 2),
                "used_gb": round(used_gb, 2),
                "available_gb": round(available_gb, 2),
                "percent": memory_percent
            },
            "process": {
                "rss_gb": round(process_rss_gb, 2),
                "vms_mb": round(process_vms / (1024**2), 2)
            },
            "alerts": generate_alerts(memory_percent, process_rss_gb)
        }

    except Exception as e:
        return {
            "error": f"Memory monitoring failed: {str(e)}",
            "timestamp": datetime.now().isoformat()
        }

def generate_alerts(system_percent, process_rss_gb):
    """Generate memory alerts based on thresholds"""
    alerts = []

    # System memory alerts
    if system_percent > 90:
        alerts.append({
            "level": "CRITICAL",
            "type": "system_memory",
            "message": f"시스템 메모리 사용량 위험 수준: {system_percent:.1f}%",
            "action": "즉시 메모리를 확보하고 Claude Code 세션을 재시작하세요"
        })
    elif system_percent > 80:
        alerts.append({
            "level": "WARNING",
            "type": "system_memory",
            "message": f"시스템 메모리 사용량 높음: {system_percent:.1f}%",
            "action": "메모리 집약적 작업을 중단하고 메모리를 정리하세요"
        })
    elif system_percent > 70:
        alerts.append({
            "level": "INFO",
            "type": "system_memory",
            "message": f"시스템 메모리 사용량 주의 필요: {system_percent:.1f}%",
            "action": "메모리 사용량을 주시하세요"
        })

    # Process memory alerts (Bun/Claude Code specific)
    if process_rss_gb > 8:  # 8GB is extremely high for a single process
        alerts.append({
            "level": "CRITICAL",
            "type": "process_memory",
            "message": f"프로세스 메모리 사용량 위험: {process_rss_gb:.2f}GB",
            "action": "즉시 Claude Code를 재시작하세요 - 메모리 누수 의심"
        })
    elif process_rss_gb > 4:  # 4GB is very high
        alerts.append({
            "level": "WARNING",
            "type": "process_memory",
            "message": f"프로세스 메모리 사용량 높음: {process_rss_gb:.2f}GB",
            "action": "메모리 누수 가능성 - 세션 재시작 고려"
        })
    elif process_rss_gb > 2:  # 2GB is concerning
        alerts.append({
            "level": "INFO",
            "type": "process_memory",
            "message": f"프로세스 메모리 사용량 관찰 필요: {process_rss_gb:.2f}GB",
            "action": "메모리 사용 패턴을 모니터링하세요"
        })

    return alerts

def main():
    """Main memory monitoring function"""
    memory_info = check_memory_usage()

    # Print memory status
    print(f"🧠 메모리 상태 모니터링 - {memory_info['timestamp']}")

    if "error" in memory_info:
        print(f"❌ {memory_info['error']}")
        return 1

    # Display system memory
    sys_mem = memory_info["system"]
    proc_mem = memory_info["process"]

    print(f"🖥️  시스템 메모리: {sys_mem['used_gb']:.1f}GB / {sys_mem['total_gb']:.1f}GB ({sys_mem['percent']:.1f}%)")
    print(f"📱 프로세스 메모리: {proc_mem['rss_gb']:.2f}GB RSS, {proc_mem['vms_mb']:.0f}MB VMS")

    # Display alerts
    alerts = memory_info["alerts"]
    if alerts:
        print("\n🚨 메모리 경고:")
        for alert in alerts:
            emoji = {"CRITICAL": "🔴", "WARNING": "🟡", "INFO": "🟢"}.get(alert["level"], "⚪")
            print(f"{emoji} {alert['level']}: {alert['message']}")
            print(f"   조치: {alert['action']}")

        # Return error code for critical alerts
        if any(alert["level"] == "CRITICAL" for alert in alerts):
            print("\n❌ 크리티컬 메모리 경고 - 세션 재시작 권장")
            return 1

    print("\n✅ 메모리 상태 정상")
    return 0

if __name__ == "__main__":
    sys.exit(main())