operations
   =         Initiate {relation: relation}
   |   SetPrimaryAttr {entityID: EntityID, attrs: [Attribute]}
;

relation
    = PrimaryWithMany { primary: shape, secondaries: [shape]}
;

shape
    = Entity {entityID: EntityID, attrs: [Attribute]}
;

attr
    = Attribute {Name: Name, Type: type}
;

type
    = Int
    | String
    | DateTime
;