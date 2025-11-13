import os
import sys
import json
import asyncio
import glob
import shutil
import copy
from typing import List, Dict, Optional
import time
from datetime import datetime

# Add project root to path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

# FastAPI imports
from fastapi import FastAPI, UploadFile, File, HTTPException, WebSocket, WebSocketDisconnect, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel
import uvicorn

from utils.logger import logger
import ast
from utils import kt_gen as constructor
from config import get_config, ConfigManager, prompt_templates
from utils import graph_processor

# # Try to import GraphRAG components
# try:
#     from utils import kt_gen as constructor
#     from config import get_config, ConfigManager, prompt_templates

#     GRAPHRAG_AVAILABLE = True
#     logger.info("✅ graph-parser-server components loaded successfully")
# except ImportError as e:
#     GRAPHRAG_AVAILABLE = False
#     logger.error(f"⚠️  graph-parser-server not available: {e}")

app = FastAPI(title="graph-parser-server Unified Interface", version="1.0.0")

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Global variables
active_connections: Dict[str, WebSocket] = {}

CONFIG = get_config()



class ConnectionManager:
    def __init__(self):
        self.active_connections: Dict[str, WebSocket] = {}

    async def connect(self, websocket: WebSocket, client_id: str):
        await websocket.accept()
        self.active_connections[client_id] = websocket

    def disconnect(self, client_id: str):
        if client_id in self.active_connections:
            del self.active_connections[client_id]

    async def send_message(self, message: dict, client_id: str):
        if client_id in self.active_connections:
            try:
                await self.active_connections[client_id].send_text(json.dumps(message))
            except Exception as e:
                logger.error(f"Error sending message to {client_id}: {e}")
                self.disconnect(client_id)


manager = ConnectionManager()


# Request/Response models
class ExtracGraphDataResponse(BaseModel):
    """ res_data 格式字段"""
    success: bool
    message: str
    graph_chunks: List[Dict] = []
    graph_vocabulary_set: set = set()
    community_reports: List[Dict] = []


# Request/Response models
class CommunityReportsResponse(BaseModel):
    """ res_data 格式字段"""
    success: bool
    message: str
    community_reports: List[Dict] = []

# Request/Response models
class RequestResponse(BaseModel):
    """ res_data 格式字段"""
    success: bool
    message: str

async def send_progress_update(client_id: str, stage: str, progress: int, message: str):
    """Send progress update via WebSocket"""
    await manager.send_message({
        "type": "progress",
        "stage": stage,
        "progress": progress,
        "message": message,
        "timestamp": datetime.now().isoformat()
    }, client_id)


@app.post("/api/extrac_graph_data", response_model=ExtracGraphDataResponse)
async def extrac_graph_data(request: Request):
    """extrac_graph_data endpoint  chunks: List[Dict], client_id: str = 'default' """
    try:
        json_request = await request.json()
        chunks = json_request["chunks"]
        user_id = json_request["user_id"]
        file_name = json_request["file_name"]
        kb_name = json_request["kb_name"]
        llm_model = json_request["llm_model"]
        llm_base_url = json_request["llm_base_url"]
        llm_api_key = json_request["llm_api_key"]
        for chunk in chunks:
            chunk["old_snippet"] = chunk["snippet"]
            chunk["snippet"] = f"{file_name.split('.')[0]}:" + chunk["snippet"]
        client_id = json_request.get("client_id", "default")
        schema = json_request.get("schema", None)
        config = get_config()
        config.construction.mode = "general"  # "agent"
        dataset = "demo"
        dataset_config = config.get_dataset_config(dataset)
        dataset_config.corpus_path = "data/demo/custom_corpus.json"
        dataset_config.schema_path = "schemas/custom.json"
        dataset_config.graph_output = "output/graphs/custom_new.json"
        config.prompts["construction"]["general"] = prompt_templates.general_zh_prompt_template
        config.construction.LLM_MODEL = llm_model
        config.construction.LLM_BASE_URL = llm_base_url
        config.construction.LLM_API_KEY = llm_api_key
        embedding_model = None
        res_data = []
        builder = constructor.KTBuilder(
            dataset,
            embedding_model,
            dataset_config.schema_path,
            schema=schema,
            mode=config.construction.mode,
            config=config
        )
        res_data = builder.build_knowledge_graph(file_name, chunks)

        # =========== 更新 graph =============
        graph_processor.update_graph(user_id, kb_name, file_name, res_data)

        # =========== 整理 graph_vocabulary_set =============
        graph_vocabulary_set = set()
        for node in builder.graph.nodes:
            node_json = builder.graph.nodes[node]
            # print(node_json)
            if node_json['properties'].get('schema_type'):
                schema_type = f"K:{node_json['properties'].get('schema_type')}"
            else:
                schema_type = "K:graph_node"
            node_msg = f"{node_json['properties']['name']}|||schema_type:{schema_type}"
            graph_vocabulary_set.add(node_msg)
        # =========== 整理 graph_vocabulary_set =============

        # =========== 整理 graph_chunks start=============
        graph_chunks = []
        for triple in res_data:
            reference_chunk_id = triple["start_node"]["properties"]["chunk id"]
            meta_data = builder.all_chunks[reference_chunk_id]["meta_data"]
            meta_data["reference_snippet"] = builder.all_chunks[reference_chunk_id]["old_snippet"]
            temp_triple = copy.deepcopy(triple)
            # 移除 start_node 中的 'chunk id'
            if 'chunk id' in temp_triple['start_node']['properties']:
                del temp_triple['start_node']['properties']['chunk id']
            # 移除 end_node 中的 'chunk id'
            if 'chunk id' in temp_triple['end_node']['properties']:
                del temp_triple['end_node']['properties']['chunk id']
            graph_data_text = f"{triple['start_node']['properties']['name']} {triple['relation']} {triple['end_node']['properties']['name']}"
            # print(graph_data_text)
            graph_chunks.append(
                {"chunk_type": "graph", "graph_data_text": graph_data_text, "graph_data": copy.deepcopy(temp_triple),
                 "meta_data": meta_data})
        # =========== 整理 graph_chunks  end =============
        # await send_progress_update(client_id, "extrac_graph_data", 10, "extrac_graph_data completed successfully!")

        return ExtracGraphDataResponse(
            success=True,
            message="Files uploaded successfully",
            graph_chunks=graph_chunks,
            graph_vocabulary_set=graph_vocabulary_set,
        )

    except Exception as e:
        # await send_progress_update(client_id, "extrac_graph_data", 0, f"Upload failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/generate_community_reports", response_model=ExtracGraphDataResponse)
async def generate_community_reports(request: Request):
    """extrac_graph_data endpoint  chunks: List[Dict], client_id: str = 'default' """
    try:
        json_request = await request.json()
        # relationships = json_request["graph_data"]
        user_id = json_request["user_id"]
        kb_name = json_request["kb_name"]
        # file_name = json_request["file_name"]
        client_id = json_request.get("client_id", "default")
        logger.info(f"generate_community_reports, user_id: {user_id}, kb_name: {kb_name}")

        # =========== 更新 graph =============
        # new_graph = graph_processor.update_graph(user_id, kb_name, file_name, relationships)
        # =========== 生成社区报告 =============
        file_path = f"./data/graph/{user_id}/{kb_name}.json"
        new_graph = graph_processor.load_graph_from_json(file_path)
        reports = graph_processor.extract_community(new_graph, "")
        await send_progress_update(client_id, "generate_community_reports", 10, "generate_community_reports completed successfully!")

        return CommunityReportsResponse(
            success=True,
            message="Files uploaded successfully",
            community_reports=reports,
        )

    except Exception as e:
        await send_progress_update(client_id, "generate_community_reports", 0, f"Upload failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/delete_file", response_model=ExtracGraphDataResponse)
async def delete_file(request: Request):
    """update graph """
    try:
        json_request = await request.json()
        user_id = json_request["user_id"]
        kb_name = json_request["kb_name"]
        file_name = json_request["file_name"]
        client_id = json_request.get("client_id", "default")

        # =========== 更新 graph =============
        graph_processor.delete_file(user_id, kb_name, file_name)
        await send_progress_update(client_id, "delete_file", 10, "delete_file completed successfully!")

        return RequestResponse(
            success=True,
            message="Files deleted successfully",
        )

    except Exception as e:
        await send_progress_update(client_id, "delete_file", 0, f"deleted failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/delete_kb", response_model=ExtracGraphDataResponse)
async def delete_kb(request: Request):
    """update graph """
    try:
        json_request = await request.json()
        user_id = json_request["user_id"]
        kb_name = json_request["kb_name"]
        client_id = json_request.get("client_id", "default")

        # =========== 更新 graph =============
        graph_processor.delete_kb(user_id, kb_name)
        await send_progress_update(client_id, "delete_kb", 10, "delete_file completed successfully!")

        return RequestResponse(
            success=True,
            message="delete_kb successfully",
        )

    except Exception as e:
        await send_progress_update(client_id, "delete_kb", 0, f"deleted failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/test_post")
async def test_post(request: Request):
    """test post"""
    json_request = await request.json()
    time.sleep(6)
    return {
        "code": 0,
        "success": True,
        "message": f"test: {json_request} successfully",
    }


@app.get("/api/test")
async def test():
    """test"""
    return {
        "code": 0,
        "success": True,
        "message": f"test successfully",
    }


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=20050)
    # chunks = [
    #     {
    #         "title": "西藏-珍贵文物全解析.docx",
    #         "meta_data": {
    #             "file_name": "西藏-珍贵文物全解析.docx",
    #             "chunk_current_num": 1,
    #             "chunk_total_num": 2
    #         },
    #         "snippet": "西藏博物馆珍贵文物全解析 灌顶国师阐化王印 文物简介：灌顶国师阐化王印为铁质，印面呈正方形，边长约 12.5 厘米 ，印钮为龙钮，龙身蜿蜒，形态矫健，仿佛随时都会腾空而起，极具威严感。印面刻有八思巴文，内容为 “灌顶国师阐化王之印”，字体线条刚劲有力，结构规整。印身整体保存较为完好，虽历经岁月洗礼，但依然能感受到其庄重肃穆的气质。 文物年份：明代 出土时间：具体出土时间无明确记载（多为传世文物） 出土地点：无明确出土地点（现藏于相关博物馆或收藏机构） 历史价值：它是明朝中央政府对西藏地方进行有效管理的重要物证，见证了明朝与西藏地区的政治关系和宗教交流。此印的存在，反映了当时明朝政府对西藏宗教领袖的册封制度，对于研究明朝的民族政策、宗教政策以及西藏地区的历史发展具有重要的意义，是研究中国古代民族关系和边疆治理的珍贵资料。 大慈法王刺绣唐卡 文物简介：大慈法王刺绣唐卡画面主要描绘了大慈法王的形象，法王面容慈祥，身着华丽的僧袍，头戴僧帽，双手结印，端坐在莲花宝座之上。唐卡背景采用了丰富的色彩和细腻的针法，绣制出祥云朵朵、花草树木以及各种佛教法器等图案，营造出庄重神圣的氛围。整幅唐卡绣工精湛，针法细腻，色彩鲜艳，历经数百年依然光彩夺目。唐卡边缘以精美的锦缎装裱，更增添了其艺术价值。 文物年份：明代 出土时间：无明确出土时间（多为传世或收藏于寺庙、博物馆） 出土地点：无明确出土地点（多流传于西藏及其他藏传佛教传播地区） 历史价值：作为藏传佛教艺术的杰出代表，大慈法王刺绣唐卡体现了明代藏传佛教艺术的高超水平和独特风格。它不仅是一件精美的艺术品，更是藏传佛教文化的重要载体，反映了当时藏传佛教的发展状况以及信徒对大慈法王的尊崇。对于研究明代藏传佛教艺术、宗教信仰、文化交流以及藏族传统刺绣工艺等方面都具有极高的价值，是了解藏传佛教文化和艺术的重要窗口。 织锦夹金五佛冠 文物简介：织锦夹金五佛冠造型独特，呈扇形，由五片莲瓣组成，每片莲瓣上分别绣有一尊佛像，佛像神态庄严，法相慈悲。五佛冠主体采用了珍贵的织锦工艺，丝线细腻，色彩绚丽，在织锦中巧妙地融入了金线，使得整个佛冠在光线的照耀下熠熠生辉，尽显华贵。冠身还装饰有各种宝石、珍珠等，进一步增添了其庄重和神圣之感。佛冠的边缘采用了精美的刺绣工艺，绣制出各种吉祥图案，如祥云、莲花等，工艺精湛，美轮美奂。 文物年份：明代 出土时间：无明确出土时间（多为传世或收藏于寺庙等宗教场所） 出土地点：无明确出土地点（多流传于藏传佛教寺庙或相关宗教文化区域） 历史价值：织锦夹金五佛冠是藏传佛教密宗的重要法器，具有极高的宗教和文化价值。它体现了明代高超的织锦和刺绣工艺水平，反映了当时藏传佛教的宗教仪式、信仰内涵以及审美观念。对于研究明代藏传佛教文化、宗教艺术、服饰文化以及工艺技术等方面都提供了珍贵的实物资料，是研究藏传佛教密宗文化和古代工艺的重要文物。 藏戏面具 文物简介：藏戏面具种类繁多，造型各异，根据不同的角色和剧情制作。常见的有国王面具、王后面具、活佛面具、神仙面具、动物面具等。面具多采用布料、皮革、木材等材质制作，经过绘画、雕刻、缝制等多道工序完成。面具色彩鲜艳，对比强烈，不同颜色代表不同的性格特征，如白色代表纯洁善良，红色代表威严勇猛，黑色代表邪恶等。面具的五官造型夸张，通过独特的设计来突出角色的特点，如国王面具的威严庄重，动物面具的生动形象等。 文物年份：藏戏面具历史悠久，其发展经历了多个时期，现存面具多为明清时期及以后制作"
    #     },
    #     {
    #         "title": "西藏-珍贵文物全解析.docx",
    #         "meta_data": {
    #             "file_name": "西藏-珍贵文物全解析.docx",
    #             "chunk_current_num": 2,
    #             "chunk_total_num": 2
    #         },
    #         "snippet": "它体现了明代高超的织锦和刺绣工艺水平，反映了当时藏传佛教的宗教仪式、信仰内涵以及审美观念。对于研究明代藏传佛教文化、宗教艺术、服饰文化以及工艺技术等方面都提供了珍贵的实物资料，是研究藏传佛教密宗文化和古代工艺的重要文物。 藏戏面具 文物简介：藏戏面具种类繁多，造型各异，根据不同的角色和剧情制作。常见的有国王面具、王后面具、活佛面具、神仙面具、动物面具等。面具多采用布料、皮革、木材等材质制作，经过绘画、雕刻、缝制等多道工序完成。面具色彩鲜艳，对比强烈，不同颜色代表不同的性格特征，如白色代表纯洁善良，红色代表威严勇猛，黑色代表邪恶等。面具的五官造型夸张，通过独特的设计来突出角色的特点，如国王面具的威严庄重，动物面具的生动形象等。 文物年份：藏戏面具历史悠久，其发展经历了多个时期，现存面具多为明清时期及以后制作 出土时间：出土时间不一，不同时期有陆续发现（部分出土于西藏及其他藏区的文化遗址、寺庙等） 出土地点：主要出土于西藏自治区以及其他藏传佛教文化传播的地区，如青海、甘肃、四川等地的藏区 历史价值：藏戏面具是藏戏艺术的重要组成部分，是藏民族文化的瑰宝。它承载着丰富的历史文化信息，反映了藏民族的宗教信仰、民俗风情、审美观念以及艺术创造力。对于研究藏戏的发展历程、藏民族的文化传承、宗教与艺术的融合等方面都具有重要意义，是研究藏文化和藏戏艺术的珍贵实物资料，也是藏民族传统文化的生动体现。"
    #     }
    # ]
    # config = get_config()
    # config.construction.mode = "general"  # "agent"
    # dataset = "demo"
    # dataset_config = config.get_dataset_config(dataset)
    # dataset_config.corpus_path = "data/demo/custom_corpus.json"
    # dataset_config.schema_path = "schemas/custom.json"
    # dataset_config.graph_output = "output/graphs/custom_new.json"
    # config.prompts["construction"]["general"] = prompt_templates.general_zh_prompt_template
    # builder = constructor.KTBuilder(
    #     dataset,
    #     None,
    #     dataset_config.schema_path,
    #     schema=None,
    #     mode=config.construction.mode,
    #     config=config
    # )
    # res_rels = builder.build_knowledge_graph("西藏-珍贵文物全解析.docx", chunks)
    # new_graph = graph_processor.update_graph("123", "西藏-珍贵文物全解析.docx", res_rels)
    # reports = graph_processor.extract_community(new_graph, ["西藏-珍贵文物全解析"])
    # print(reports)
    # graph_processor.delete_file("123", "西藏-珍贵文物全解析.docx")
