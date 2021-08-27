CREATE TABLE SearchRequest (
	query TEXT NOT NULL ,
	page_number INT NOT NULL ,
	result_per_page INT NOT NULL 
);

CREATE TABLE User (
	id INT NOT NULL ,
	username TEXT NOT NULL ,
	age INT NULL ,
	sgender ENUM("MALE","FEMALE","OTHER") NOT NULL ,
	s JSON NOT NULL ,
	stamps JSON NOT NULL 
);