
from typing import Any,Mapping,List,Tuple
from google.protobuf import json_format
ENUMDICT__Foo_User_Gender = {
	0:"MALE",
	1:"FEMALE",
	2:"OTHER"
}
def getSearchRequestColumnNames() -> List[str]:
	return "query,page_number,result_per_page,PROTO_BINARY"
	
# convert proto message class variable to INSERT-ready dictionary
def convSearchRequestProtoClassToData(value) -> Tuple:
	return (value.query,value.page_number,value.result_per_page,value.SerializeToString())
		


def getUserColumnNames() -> List[str]:
	return "id,username,Age,sgender,s,stamps,PROTO_BINARY"
	
# convert proto message class variable to INSERT-ready dictionary
def convUserProtoClassToData(value) -> Tuple:
	return (value.id,value.username,value.Age if value.HasField("Age") else None,ENUMDICT__Foo_User_Gender[value.sgender],value.s,json_format.MessageToJson(value.stamps),value.SerializeToString())
		