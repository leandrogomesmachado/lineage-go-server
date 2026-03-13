// Script de criação da collection characters para GameServer
// Baseado em: raptors_datapack/sql/characters.sql

db = db.getSiblingDB('l2raptors');

// Criar collection characters
db.createCollection('characters', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['account_name', 'obj_id', 'char_name', 'level', 'race', 'classid', 'sex'],
      properties: {
        account_name: {
          bsonType: 'string',
          description: 'Nome da conta do jogador'
        },
        obj_id: {
          bsonType: 'int',
          description: 'ID unico do personagem'
        },
        char_name: {
          bsonType: 'string',
          maxLength: 35,
          description: 'Nome do personagem'
        },
        level: {
          bsonType: 'int',
          minimum: 1,
          maximum: 80,
          description: 'Nivel do personagem'
        },
        maxHp: { bsonType: 'int' },
        curHp: { bsonType: 'int' },
        maxCp: { bsonType: 'int' },
        curCp: { bsonType: 'int' },
        maxMp: { bsonType: 'int' },
        curMp: { bsonType: 'int' },
        face: { bsonType: 'int' },
        hairStyle: { bsonType: 'int' },
        hairColor: { bsonType: 'int' },
        sex: { bsonType: 'int', minimum: 0, maximum: 1 },
        heading: { bsonType: 'int' },
        x: { bsonType: 'int' },
        y: { bsonType: 'int' },
        z: { bsonType: 'int' },
        exp: { bsonType: 'long', minimum: 0 },
        expBeforeDeath: { bsonType: 'long', minimum: 0 },
        sp: { bsonType: 'int', minimum: 0 },
        karma: { bsonType: 'int' },
        pvpkills: { bsonType: 'int', minimum: 0 },
        pkkills: { bsonType: 'int', minimum: 0 },
        clanid: { bsonType: 'int' },
        race: { 
          bsonType: 'int',
          minimum: 0,
          maximum: 5,
          description: '0=Human, 1=Elf, 2=DarkElf, 3=Orc, 4=Dwarf'
        },
        classid: { 
          bsonType: 'int',
          description: 'ID da classe do personagem'
        },
        base_class: { bsonType: 'int' },
        deletetime: { bsonType: 'long' },
        cancraft: { bsonType: 'int' },
        title: { bsonType: 'string' },
        accesslevel: { bsonType: 'int' },
        online: { bsonType: 'int', minimum: 0, maximum: 1 },
        onlinetime: { bsonType: 'int' },
        char_slot: { bsonType: 'int' },
        lastAccess: { bsonType: 'long' },
        clan_privs: { bsonType: 'int' },
        wantspeace: { bsonType: 'int' },
        isin7sdungeon: { bsonType: 'int' },
        punish_level: { bsonType: 'int' },
        punish_timer: { bsonType: 'int' },
        power_grade: { bsonType: 'int' },
        nobless: { bsonType: 'int' },
        subpledge: { bsonType: 'int' },
        last_recom_date: { bsonType: 'long' },
        lvl_joined_academy: { bsonType: 'int' },
        apprentice: { bsonType: 'int' },
        sponsor: { bsonType: 'int' },
        varka_ketra_ally: { bsonType: 'int' },
        clan_join_expiry_time: { bsonType: 'long' },
        clan_create_expiry_time: { bsonType: 'long' },
        death_penalty_level: { bsonType: 'int' },
        created_at: { bsonType: 'date' },
        updated_at: { bsonType: 'date' }
      }
    }
  }
});

// Criar indices
db.characters.createIndex({ obj_id: 1 }, { unique: true });
db.characters.createIndex({ account_name: 1 });
db.characters.createIndex({ char_name: 1 }, { unique: true });
db.characters.createIndex({ clanid: 1 });
db.characters.createIndex({ online: 1 });
db.characters.createIndex({ level: -1 });

print('Collection characters criada com sucesso!');
print('Indices criados: obj_id (unique), account_name, char_name (unique), clanid, online, level');
