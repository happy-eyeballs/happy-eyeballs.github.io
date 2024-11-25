import argparse
import datetime
from flask import Flask, request, jsonify
from werkzeug.middleware.proxy_fix import ProxyFix
import logging
import json
import os

app = Flask(__name__)

DATA_FILE = '{date}-results.jsonl'
V2DATA_FILE = '{date}-v2results.jsonl'
DNSURSERDATA_FILE = '{date}-dns-user-info.jsonl'
DNSDATA_FILE = '{date}-dns-results.jsonl'
OUTPUT_DIR = '{{ upload_dir }}'
V2OUTPUT_DIR = '{{ v2_oupload_dir }}'
DNSOUTPUT_DIR = '{{ dns_upload_dir }}'

@app.route('/results', methods=['POST'])
def upload_data():
    # Get JSON data from request
    try:
        data = request.get_json()

        if not isinstance(data, list):
            return 'FAILURE', 500
        if 'runCount' not in data[0] or 'id' not in data[0]:
            return 'FAILURE', 500

        date = datetime.datetime.now()
        date_str = date.strftime('%Y-%m-%d')
        month_dir = date.strftime('%Y/%m')
        data_filepath = os.path.join(OUTPUT_DIR, month_dir, DATA_FILE.format(date=date_str))

        os.makedirs(os.path.join(OUTPUT_DIR, month_dir), exist_ok=True)

        with open(data_filepath, 'a') as f:
            json.dump(data, f)
            f.write('\n')

        return jsonify({"message": "Data uploaded successfully"}), 200
    except Exception:
        return 'FAILURE', 500


@app.route('/v2results', methods=['POST'])
def upload_v2data():
    # Get JSON data from request
    try:
        data = request.get_json()

        if not isinstance(data, list):
            return 'FAILURE', 500
        if 'runCount' not in data[0] or 'id' not in data[0]:
            return 'FAILURE', 500

        date = datetime.datetime.now()
        date_str = date.strftime('%Y-%m-%d')
        month_dir = date.strftime('%Y/%m')
        data_filepath = os.path.join(V2OUTPUT_DIR, month_dir, V2DATA_FILE.format(date=date_str))

        os.makedirs(os.path.join(V2OUTPUT_DIR, month_dir), exist_ok=True)

        with open(data_filepath, 'a') as f:
            json.dump(data, f)
            f.write('\n')

        return jsonify({"message": "Data uploaded successfully"}), 200
    except Exception:
        return 'FAILURE', 500


@app.route('/dnsresults', methods=['POST'])
def upload_dnsdata():
    # Get JSON data from request
    try:
        data = request.get_json()

        if not isinstance(data, list):
            return 'FAILURE', 500
        if 'runCount' not in data[0] or 'id' not in data[0]:
            return 'FAILURE', 500

        date = datetime.datetime.now()
        date_str = date.strftime('%Y-%m-%d')
        month_dir = date.strftime('%Y/%m')
        data_filepath = os.path.join(OUTPUT_DIR, month_dir, DNSURSERDATA_FILE.format(date=date_str))

        os.makedirs(os.path.join(OUTPUT_DIR, month_dir), exist_ok=True)

        with open(data_filepath, 'a') as f:
            json.dump(data, f)
            f.write('\n')

        return jsonify({"message": "Data uploaded successfully"}), 200
    except Exception:
        return 'FAILURE', 500


@app.route('/dns-query', methods=['POST'])
def log_dns_query():
    # Get JSON data from request
    try:
        data = request.get_json()

        if not isinstance(data, dict):
            return 'FAILURE', 501
        if 'ns_ip' not in data:
            return 'FAILURE', 502

        date = datetime.datetime.now()
        date_str = date.strftime('%Y-%m-%d')
        month_dir = date.strftime('%Y/%m')
        data_filepath = os.path.join(DNSOUTPUT_DIR, month_dir, DNSDATA_FILE.format(date=date_str))

        os.makedirs(os.path.join(DNSOUTPUT_DIR, month_dir), exist_ok=True)

        with open(data_filepath, 'a') as f:
            json.dump(data, f)
            f.write('\n')

        return jsonify({"message": "Data uploaded successfully"}), 200
    except Exception as e:
        logging.error(e)
        return 'FAILURE', 503


# if __name__ == '__main__':
#     parser = argparse.ArgumentParser('Results Upload Handler')
#     parser.add_argument('-o', '--output-directory', required=True, type=str, help='Directory where results are stored')
#     parser.add_argument('--v2-output-directory', required=True, type=str, help='Directory where v2 results are stored')
#     parser.add_argument('-d', '--dns-output-directory', required=True, type=str, help='Directory where DNS results are stored')
#     args = parser.parse_args()
#     OUTPUT_DIR = args.output_directory
#     V2OUTPUT_DIR = args.v2_output_directory
#     DNSOUTPUT_DIR = args.dns_output_directory


#     app.run(host='127.0.0.1', port=40_000)
